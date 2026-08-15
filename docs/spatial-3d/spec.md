# 3D spatial partitioning — Living Spec

**Linear:** [PMP-572](https://linear.app/pixel-master-pro/issue/PMP-572/3d-spatial-partitioning-volume-queries-and-ordered-ray-traversal)

**Status:** not started.

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

- [ ] **Phase 1 — the structure and volume queries**: insert an item with a
      `Box3` bound; query all items overlapping a given `Box3` or `Sphere`;
      clear and rebuild reusing storage. Benchmark the rebuild at the scale the
      consumer cares about (5 000 items is the number `pickindex.go` measures at,
      so the comparison is direct).
- [ ] **Phase 2 — ordered ray traversal**: visit cells strictly near-to-far along
      a `Ray3`, yielding candidates, so a caller that has confirmed a hit can
      stop as soon as the next cell cannot contain anything nearer. **The
      ordering is the feature** — an unordered ray query is just a volume query
      against the ray's bounding box and is not worth building. Handle the ray
      originating inside the structure, outside it, and missing it entirely.
- [ ] **Phase 3 — the pairs sweep**: candidate-pair generation for a broad phase
      (every pair of items sharing a cell, each pair yielded once). The 2D
      `SimpleHashGrid` does this via `IterateCells` + `IterateNextNeighbors`;
      the 3D version has 26 neighbours rather than 8, and the "next" half of
      them is 13, so the de-duplication rule needs restating rather than
      translating.
- [ ] **Phase 4 — document the height decision**: whether this structure
      subdivides all three axes or follows `pickindex.go` in treating the world
      as a ground plane. See Open questions; this phase is where the answer and
      its measurement land.

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

- **Grid or octree** — decided as "grid" above on the strength of the access
  pattern, but Phase 1's benchmark is what should confirm it.
- **Does it subdivide height?** `pickindex.go` deliberately does not, arguing the
  world is a ground plane and *"a ray passing high over a crowd still collects
  the candidates beneath it and rejects them in the narrow phase, which costs a
  sphere test each and is the price of not maintaining a third axis almost
  nothing uses."* That reasoning was written for picking in a specific scene. A
  collision broad phase in a world with bridges or flying units has a different
  answer. **Do not settle this by argument** — build it with the axis, measure a
  scene with vertical spread, and record the number. A configurable cell count
  per axis (with 1 in Y being the flat case) may make the question disappear.
- Whether the structure should be usable as the pick index directly, or whether
  picking keeps its own. That is PMP-578's call, not this one's; this task's job
  is to make the choice available.

## Revision history

- v1 (2026-08-14) — initial spec.
