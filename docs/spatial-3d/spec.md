# 3D spatial partitioning — Living Spec

**Linear:** [PMP-572](https://linear.app/pixel-master-pro/issue/PMP-572/3d-spatial-partitioning-volume-queries-and-ordered-ray-traversal)

**Status:** phases 1-4 built in `spatial/grid3.go`. Not yet consumed by
pixelmaster — PMP-578 and PMP-581 are what pick it up.

## Goal

Give `gomath/spatial` a 3D structure that answers the two questions a 3D game
asks of it: **which items overlap this volume** (the collision broad phase) and
**what does this ray hit first** (picking, hitscan, line of sight) — the second
with cells visited in near-to-far order so the search can stop early.

## Depends on

Nothing in gomath. Can start immediately, in parallel with PMP-570.

## Depended on by

- **PMP-578** (pixelmaster `Spatial3D`) — the broad phase is built on this.
- **PMP-581** (pixelmaster 3D queries) — the ray traversal is what makes a cast
  cheap, and PMP-581 also decides whether `render3d`'s bespoke index folds into
  the consumer of this.

## The gap, and the evidence it is real

`spatial/` contains `Quadtree`, `BoxedQuadtree`, `HashGrid`,
`SimpleHashGrid`, `UnlimitedHashGrid` — **all 2D**. There is no octree, no 3D
grid, and no ray query of any kind at any dimension.

The strongest evidence that this needs building is that a consumer already hit
it and worked around it. `../pixelmaster/render3d/pickindex.go` opens with a
comment headed *"Why a purpose-built grid rather than gomath/spatial"*:

> That package has quadtrees and hash grids, and none of them answers the
> question asked here: every query it offers is a box, a point or a neighbour
> sweep, and there is no ray traversal anywhere in it. Its hash grid also wraps
> toroidally, which for a ray is not merely unhelpful but wrong — a ray leaving
> one edge would collect candidates from the far side of the world.

**Read that file before designing this.** It is a working implementation of half
of what this task wants, it is in production use for entity picking, and it
carries two measured results that should not be rediscovered the hard way:

- A **counting sort into one flat array was slower** than the reused per-cell
  slices — ~595 µs against 411 µs for five thousand entities. It walks each
  candidate's cells twice, and the allocations it was meant to avoid were not
  happening anyway. The comment's own verdict: *"The obvious optimisation was
  the wrong one."*
- Composing a transform matrix **per entity** during rebuild cost more than the
  linear scan the index replaced. It keeps transforms by value and composes only
  for candidates that survive the cheap sphere test.

It also documents a deliberate limitation this task must decide whether to
inherit: the grid divides X and Z and **ignores height entirely**, on the
grounds that the world is a ground plane with things standing on it.

## Non-goals

- Replacing or generifying the 2D structures. They run the shipped game.
- A BVH, or anything with a build cost that assumes static geometry. The
  consumer rebuilds every tick from moving transforms.
  **Corrected in v3** — that reasoning is about the *dynamic* layer and does not
  transfer to a static one, where "assumes static geometry" is a description of
  the case rather than an objection to it. See Open questions.
- Persistent incremental update (insert/remove/relocate as the world changes).
  The consumer's pattern is clear-and-rebuild; if incremental update is wanted
  later it is its own task, with its own evidence.
- Concurrency. The 2D structures document their thread-safety story; this one
  starts single-threaded and says so.

## Assumptions & constraints

- **Rebuild-per-tick is the access pattern.** Optimise for a cheap full rebuild
  with reused backing storage, not for a cheap single insert. `pickindex.go`
  truncates its cell slices rather than discarding them, so a steady scene
  allocates nothing after the first frame — that is the target behaviour.
- **No toroidal wrapping.** Out-of-bounds items clamp or are rejected
  explicitly. Wrapping is what disqualified `SimpleHashGrid` for ray use.
- Generic over `T Float` and an item ID type, matching the existing structures'
  shape (`BoxedQuadtree[T, ID]`).
- `float32` is the real consumer's instantiation.
- Pinned benchmarks only (`taskset -c 0 env GOMAXPROCS=1 ... -count=8`, medians).

## Affected areas

- New: `spatial/grid3.go` (or `octree.go` — see Open questions) and its tests.
- New: `spatial/ray3.go` or a traversal method on the structure.
- `spatial/types.go` — an item interface for 3D-bounded items, sibling to
  `BoundedItem`/`BoxedItem`.
- `CLAUDE.md` — Project Structure's `spatial/` line.

## Phases

- [x] **Phase 1 — the structure and volume queries**: insert an item with a
      `Box3` bound; query all items overlapping a given `Box3` or `Sphere`;
      clear and rebuild reusing storage. Benchmark the rebuild at the scale the
      consumer cares about (5 000 items is the number `pickindex.go` measures at,
      so the comparison is direct).
- [x] **Phase 2 — ordered ray traversal**: visit cells strictly near-to-far along
      a `Ray3`, yielding candidates, so a caller that has confirmed a hit can
      stop as soon as the next cell cannot contain anything nearer. **The
      ordering is the feature** — an unordered ray query is just a volume query
      against the ray's bounding box and is not worth building. Handle the ray
      originating inside the structure, outside it, and missing it entirely.
- [x] **Phase 3 — the pairs sweep**: candidate-pair generation for a broad phase
      (every pair of items sharing a cell, each pair yielded once). The 2D
      `SimpleHashGrid` does this via `IterateCells` + `IterateNextNeighbors`;
      the 3D version has 26 neighbours rather than 8, and the "next" half of
      them is 13, so the de-duplication rule needs restating rather than
      translating.
- [x] **Phase 4 — document the height decision**: whether this structure
      subdivides all three axes or follows `pickindex.go` in treating the world
      as a ground plane. See Open questions; this phase is where the answer and
      its measurement land.

## What was built

`spatial/grid3.go` — `Grid3[T Float, ID comparable]`, a uniform grid with a
per-axis cell size.

- `NewGrid3(cellSize Vector3)`, `NewUniformGrid3(cellSize)`,
  `NewGroundGrid3(cellSize)`. A cell size of zero on an axis leaves that axis
  unsubdivided, which is how the ground-plane case is expressed — the
  configurable-count-per-axis idea from Open questions, and it did make the
  height question dissolve into a constructor choice.
- `Add` / `Build` / `Clear`. Queries build lazily if items were added since the
  last `Build`, so forgetting the call costs a rebuild rather than silently
  returning nothing.
- `QueryBox`, `QuerySphere` — `iter.Seq[ID]`, exact overlap test, each item once.
- `RayCandidates` — `iter.Seq2[T, ID]` of (cell entry distance, item).
- `Pairs` — `iter.Seq2[ID, ID]`, every pair sharing a cell, once.
- No `Remove` or `Relocate`, per the non-goals.

### De-duplication without a visited set

An item is filed into every cell its bound covers, so a query walking a range of
cells meets it several times. The rule that yields it once: an item is yielded
only from the **componentwise maximum of its own lowest cell and the query's
lowest cell**, which is the lowest cell of the intersection of two boxes and so
is guaranteed to be in the walk. Constant time per candidate, no state between
cells, no allocation. `Pairs` uses the same rule against the other item's lowest
cell rather than the query's.

This is why the 26-neighbour restatement Phase 3 anticipated was not needed at
all. `SimpleHashGrid` sweeps neighbours because it files an item into the one
cell holding its *position*; a grid that files by *bound* has already accounted
for them, since two overlapping bounds necessarily meet inside some cell.

### Measurements

Pinned (`taskset -c 0 env GOMAXPROCS=1`, medians of 7), 5 000 items, cell size 4,
a 200-unit world, `float32`. "flat" is a ground plane, "spread" has as much
vertical extent as horizontal. Benchmarks live in `spatial/benchmarkGrid3_test.go`.

| | rebuild | ray cast | candidates/cast | pairs sweep | pairs yielded |
|---|---|---|---|---|---|
| flat, Y subdivided | 860 µs | 0.96 µs | 6.8 | 928 µs | 17 317 |
| flat, Y flat | 598 µs | 8.0 µs | 191 | 578 µs | 24 660 |
| spread, Y subdivided | 945 µs | 2.2 µs | 5.0 | 363 µs | 1 118 |
| spread, Y flat | 600 µs | 5.9 µs | 136 | 576 µs | 24 660 |

The linear scan the ray replaces is **151 µs** per cast, so the traversal is
70-160x faster than testing every item, which is the claim the ordering was
built to support. Rebuilds allocate **nothing** after the first.

Against `pickindex`, re-measured on this machine rather than taken from its
comment (`BenchmarkPickIndexRebuild`, `BenchmarkPickIndexed`, same 5 000
entities, pinned, medians of 7): rebuild **444 µs**, cast **12.5 µs**. So
`Grid3`'s ground-plane rebuild (598 µs) is the same order and its cast is
faster either way — 8.0 µs flat, 0.96 µs subdivided.

**That is not a controlled comparison and should not be quoted as a ratio.**
The scenes differ (spread 120 against a 200-unit world) and so does the work:
`pickindex`'s rebuild also walks the ECS and derives a bounding sphere per
entity, which `Grid3` never does because its caller does it before calling
`Add`; while `Grid3`'s number includes 5 000 `Add` calls (23 µs, measured
separately). The item spans differ too, and filing cost follows span count.
What the numbers support is "the same order of magnitude on the same
machine", nothing finer. A real head-to-head would need one scene driving
both, which has not been built.

Three things were measured rather than assumed, and two of them changed the code:

- **Clearing only the occupied cells, not the grid.** In 3D the grid is mostly
  empty, and sweeping every cell to truncate it was the largest single cost in a
  rebuild. Keeping a list of the cells that hold anything bounds the clear by the
  item count instead of the cell count. On a dense scene the two are within the
  noise of a loaded machine and it was not possible to separate them; on a sparse
  one (250 047 cells, 441 occupied) it is **548 µs against 36 µs, 15x**. The
  choice was made on that worst case, not on a median.
- **A reciprocal instead of a divide in `cellOf`, and a narrower parallel
  array.** Every item pays `cellOf` twice per rebuild and it was half the profile.
  Storing only each item's *lowest* cell rather than its whole span halved that
  array — the high corner was never read, because the walk already knows the item
  is in the cell it is looking at. Interleaved A/B: **1.2-1.6x** on rebuild, no
  case slower.
- **`Pairs` walking the occupancy list rather than the grid.** 2 211 µs to 363 µs
  on the spread scene, **6.1x**.

### The cell size does not transfer from 2D

"Cells about the size of the things in them" is the 2D rule and it produces a
grid that is mostly empty in 3D, because halving the cell size octuples the cell
count. Measured on the spread scene:

| cell size | cells | occupancy | rebuild |
|---|---|---|---|
| 2 | 132 651 (clamped from 1 061 208) | 15 % | 933 µs |
| 4 | 132 651 | 15 % | 932 µs |
| 8 | 17 576 | 47 % | 721 µs |
| 16 | 2 197 | 96 % | 532 µs |
| 32 | 343 | 99 % | 435 µs |

Cell size 2 and cell size 4 are the same row because `MaxGrid3Cells` doubled the
former until it fit — the guard firing, visibly. A 3D consumer should expect to
want a cell rather larger than its items, and `OccupiedCells` is the number to
look at when choosing.

## Design notes

**Uniform grid or octree.** `pickindex.go` chose a uniform grid and its
reasoning transfers: cell size ≈ the size of the things being indexed, cheap
rebuild, no tree to rebalance, and DDA traversal for the ray. An octree adapts
better to wildly non-uniform density and costs more to rebuild every tick.
Given the access pattern is rebuild-per-tick, the grid is the better default and
the octree needs evidence before it is worth building. Start with the grid.

**The ray traversal is a 3D DDA.** Step along whichever axis has the nearest
next cell boundary. `pickindex.go`'s `axisStep` and `walk` are the 2D-in-XZ form
of exactly this; extending to three axes is mechanical, and its `entryDistance`
already handles the ray-starts-outside case.

**What "early exit" requires of the caller.** The traversal yields candidates
per cell and cannot itself know whether a candidate is a real hit — that is a
narrow-phase question. So the API must let the caller say *"I have a hit at
distance d, stop"*. An iterator the caller can break out of is enough, and Go's
range-over-func gives that for free; the 2D structures already return iterators,
so this matches.

## Open questions

- **Grid or octree** — settled as the grid. A rebuild of 5 000 items is under a
  millisecond and allocates nothing; there is no evidence for paying a tree's
  build cost against an access pattern that throws the structure away each tick.
  An octree still needs its own evidence before it is worth building.

  **v3 narrows what that evidence would have to be**, because the paragraph
  above only argues about the *dynamic* layer and the question was raised again
  against the static one.

  The argument as written rejects the octree on **density adaptation**, which is
  the weakest of the three things a tree offers. The two it does not address:

  - **Persistence and mutation.** `Grid3` is `Add` / `Build` / `Clear`, batch
    only, by design. `Quadtree` has `Insert`, `Remove`, `Update`, `Expand`,
    `Collapse`. That is a difference of lifecycle, not of dimension, and nothing
    in the measurements speaks to it.
  - **Large-span items**, which is the real gap. An item is filed into every
    cell its bound covers, so filing cost follows span count. One building or
    cliff face in a grid sized for characters files into thousands of cells, and
    `MaxGrid3Cells` responds by doubling the cell size until it fits — degrading
    the grid *for everything else*. A tree stops a large item at a shallow node:
    one entry. This is the classic grid weakness and `Grid3` does not address it.

  **So the condition is narrow: a static layer with a wide size range.** Not
  merely static, and not merely dense. `pixelmaster`'s 2D `simulation.Spatial`
  has exactly this shape already — a `BoxedQuadtree` for props and regions,
  inserted at level load and never rebuilt, beside a per-tick hash grid for
  actors — and the 3D side has no equivalent of the static half.

  **This is measurable before anything is built, with what already exists.**
  Load the static scenery into a `Grid3` and compare total cell entries against
  `Len()`; `OccupiedCells()` and `CellCounts()` are exposed for it. If a handful
  of large items generate most of the entries, that is the evidence, in numbers.
  If they do not, there is nothing to build. **Do not build the tree first and
  measure afterwards** — the quadtree's own history is the argument for that
  order (see below).

  **And if the evidence does arrive, an octree is probably still the wrong
  answer.** For a layer that is genuinely static, a BVH beats an octree — it is
  built once over known geometry, adapts to the items rather than to space, and
  has no mutation paths at all. The Non-goals entry excluding it was reasoning
  about the dynamic layer. Cheaper than either: hold oversized items in a
  separate coarse list checked on every query, or a two-level grid. That targets
  the large-span problem specifically and needs none of a tree's machinery.

  **What the 2D trees' age actually tells us.** They date from December 2025 and
  needed a bug-fix campaign in July 2026 — *"fix iterator yield inversion, edge
  panics, and UpdateItem ghosts"*, *"fix Quadtree.Depth to follow the deepest
  quadrant"*, *"remove dead Quadtree.allItems index, make Update honest"*,
  *"BoxedQuadtree.Remove and superseded-type bugs"*. Six commits, and nearly all
  of them in the **mutable** paths. An octree is that machinery with eight
  children rather than four and a third axis of edge cases. Their existence is a
  cost estimate, not a template arguing that the 3D sibling should exist for
  symmetry; `Grid3` avoids that entire bug class by being immutable between
  builds, which is a large part of why it landed clean.
- **Does it subdivide height?** — **Answered: subdivide by default, and keep the
  flat case one constructor away.** The measurement is the table above.

  `pickindex.go` is right about the mechanism and wrong about the magnitude. Its
  argument is that a ray over a crowd "collects the candidates beneath it and
  rejects them in the narrow phase, which costs a sphere test each". At 5 000
  entities that is **191 tests per cast, not a handful**, and subdividing height
  makes the cast **8.4x faster on the ground-plane scene** — the very scene the
  argument was written for. It is not the vertical spread that decides it; a flat
  world with enough things in it is already dense enough for the axis to pay.

  What keeps the flat case worth having is the rebuild. Subdividing height costs
  260-345 µs per rebuild and saves 3.7-7.1 µs per cast, so it breaks even at
  roughly **37 casts per frame on the flat scene and 93 on the spread one**.
  Picking is *one* cast per frame, so `pickindex`'s choice is right for
  `pickindex` — and would be wrong for line of sight over a hundred units, or for
  a broad phase, where the spread scene's pairs sweep is both faster *and* yields
  22x fewer candidate pairs with the axis than without.

  So the answer is a property of the query load, not of the world, which is
  exactly why it belongs in a constructor argument rather than in the structure.
- Whether the structure should be usable as the pick index directly, or whether
  picking keeps its own. Still PMP-578's call; the choice is now available, and
  `NewGroundGrid3` is `pickindex`'s shape exactly.

## A bug found in the consumer while reading it

`pickindex.go`'s `walk` reports its cell entry distance measured from where the
ray **enters the grid**, not from the ray's origin: `axisStep` is handed
`ray.At(entry)` and returns `(boundary - position) / direction`, and `cellEntry`
is then set from that without adding `entry` back. A caller comparing a
confirmed hit — a true distance from the origin — against that number is
comparing two different origins, and the entry distance is always the smaller.

It is a performance bug, not a correctness one: under-reporting the cell entry
makes the early exit fire later than it could, never earlier, so picking still
returns the right entity. `Grid3` shifts the axis distances by the entry
distance at setup and pins the units with a test
(`TestGrid3RayDistanceIsFromTheOrigin`). Worth fixing in `pickindex` if it is
kept rather than folded into this — PMP-581's call.

## Revision history

- v3 (2026-08-15) — the octree question reopened against the *static* layer and
  narrowed rather than reversed: the grid stands, the condition that would
  justify a tree is a static layer with a wide size range, and the measurement
  that settles it can be run against `Grid3` today. Corrected the BVH non-goal,
  which had excluded it by reasoning about the dynamic layer. No code changed.
- v2 (2026-08-15) — built. All four phases landed in `spatial/grid3.go`; results,
  the height answer, and the cell-size finding recorded above.
- v1 (2026-08-14) — initial spec.
