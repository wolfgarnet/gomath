# Capsule3 — Living Spec

**Linear:** [PMP-571](https://linear.app/pixel-master-pro/issue/PMP-571/capsule3-the-primitive-a-3d-character-collider-needs)

**Status:** complete (2026-08-24). All four phases landed; `go test ./...` green,
pixelmaster and gorendr3d still build.

## Goal

Add `Capsule3[T]` to gomath — the type, its intersection and penetration set
against the existing 3D volumes, and a ray cast — so that pixelmaster's 3D
collider vocabulary (PMP-576) can model a walking character with the shape the
job actually wants.

## Depends on

- **[PMP-570](https://linear.app/pixel-master-pro/issue/PMP-570/complete-the-3d-penetration-matrix-and-a-consistency-harness-to-hold)** — *3D penetration matrix*.
  Read `docs/3d-penetration-matrix/spec.md` before starting. Two things come from
  it and must not be re-derived here: the exact function signature shape for a
  penetration test, and the consistency harness this task's new pairs are added
  to as rows. Its Design notes section states the normal convention in full.

## Depended on by

- **PMP-573** (swept volumes) — a swept sphere *is* a capsule; that identity is
  most of why PMP-573 is small.
- **PMP-574** (gorendr3d mesh proxies) — the capsule fit returns this type.
- **PMP-575** (gorendr3d debug wireframes) — draws this type.
- **PMP-576** (pixelmaster `Collider3D`) — the shape vocabulary includes it.

## Why a capsule and not something already here

A capsule is a segment with a radius: every point within distance `r` of a line
segment. For anything that walks it beats both alternatives already in the
library.

- **Against a sphere**: a sphere wide enough to contain a character's shoulders
  is also that wide at the feet, so the character cannot get near a wall. A
  capsule is the right width at every height.
- **Against an OBB**: a box has corners and edges, and corners catch. Walking a
  box along a wall snags on every seam in the geometry. A capsule has no
  features to catch on — its surface is everywhere smooth — which is why it
  slides over small steps rather than stopping dead at them.
- **Cost**: `OBB3OBB3` is the most expensive test in the library, fifteen
  separating axes. Capsule-vs-anything reduces to a **distance from a segment**,
  which is a handful of dot products. The most common entity in the game should
  not be paying for the most expensive test.

## Non-goals

- **Rotating the capsule as a first-class concern.** A character capsule is
  upright; store the segment in world space and the rotation is implicit in the
  two endpoints. No separate orientation field.
- Capsule-vs-capsule *continuous* tests. PMP-573.
- A cylinder. It looks similar and is much worse: flat caps reintroduce the
  edges the capsule exists to avoid, and the tests are far uglier. If something
  genuinely needs a cylinder later it is its own task.
- 2D capsules (a stadium). Nothing has asked.

## Assumptions & constraints

- Everything from PMP-570's *Assumptions & constraints* applies unchanged:
  the shape1 → shape2 normal convention, `float32`-viable tolerances, value
  receivers, `collider` imports `gomath` and never the reverse, pinned
  benchmarks.
- **Store the segment, not centre-plus-height.** `Segment` (two `Vector3`) plus
  `Radius` makes the degenerate case — zero-length segment — exactly a sphere,
  which keeps the tests honest by reduction. A centre-plus-half-height-plus-axis
  form has to special-case a zero axis instead.
- The type lives at the root (`capsule3.go`); its cross-shape collision lives in
  `collider/` (import cycle, per the repo rule).

## Affected areas

- New: `capsule3.go`, `capsule3_test.go` at the root.
- New: `collider/capsule3.go`, `collider/capsule3_test.go`.
- `collider/ray3.go` — gains `Ray3Capsule3`.
- The 3D consistency table from PMP-570 — new rows.
- `CLAUDE.md` — Project Structure's 3D volumes line.

## Phases

- [x] **Phase 1 — the type**: `Capsule3[T]` with `Segment gomath.LineSegment3[T]`
      and `Radius T`. (The type question is settled under Resolved questions.) Constructors
      for the upright character case (`NewCapsule3Upright(base, height, radius)`)
      and the general one. `BoundingBox() Box3[T]`, `Center()`, `ClosestPoint()`,
      the self-contained helpers the root types carry. **A segment-point distance
      helper is the whole engine of this feature** — if `distance.go` has no 3D
      segment-point routine, it gets one here.
- [x] **Phase 2 — sphere and capsule pairs**: `Capsule3Sphere`,
      `Capsule3Capsule3`, and both penetrations. These are the cheap analytic
      ones: sphere is point-to-segment distance against `r1+r2`, capsule-capsule
      is segment-to-segment distance against the same. Land them first because
      they are exact and make good oracles for what follows.
- [x] **Phase 3 — box pairs**: `Capsule3Box3`, `Capsule3OBB3` and their
      penetrations. Segment-vs-box is genuinely harder than the above and is
      where an approximation will be tempting. Whatever is chosen, say in this
      document what it is and where it is wrong — a capsule resting on a box
      corner is the case to name.
- [x] **Phase 4 — `Ray3Capsule3`**, and the consistency + oracle rows for every
      pair added above, following PMP-570's harness.

## Design notes

**The reduction that keeps this honest.** A capsule with a zero-length segment
is exactly a sphere of the same radius. Every capsule test should therefore
agree with its sphere counterpart when the segment degenerates — same
penetration, same normal, same boolean. That is a free oracle for the entire
feature and the repo already uses the technique (CLAUDE.md: *"reduction to a
simpler shape (e.g. circular ellipse == circle)"*).

A second free oracle: a capsule and a sphere overlap iff the distance from the
sphere's centre to the capsule's segment is less than `r1 + r2`. That is one
line, computable independently of whatever the real implementation does.

**Where the normal comes from.** For every capsule pair the MTV direction is the
direction of closest approach between the capsule's *segment* and the other
shape — not between centres. Using centres gives a plausible-looking normal
that is wrong whenever the contact is near an end cap, which is exactly where a
character touches the ground.

**The parallel-segments degeneracy.** Segment-to-segment distance has no unique
closest pair when the segments are parallel and overlapping in projection — two
capsules lying side by side, which for a crowd of standing characters is the
common case, not the exotic one. Pick the midpoint of the overlapping interval
and pin it in a test. `collider/obb3.go` records the analogous lesson for
parallel boxes and its `TestOBB3ParallelBoxesOverlap` exists for that reason;
this is the same failure in a different shape.

## Resolved questions

- **`LineSegment3[T]` exists** — the type earned itself. The segment-distance
  routines are the engine of the whole feature and wanted a home with a test file
  of its own rather than being hidden inside the capsule, and PMP-573 and PMP-575
  both talk about 3D segments. Fields are `P1`/`P2`, matching the 2D
  `LineSegment` so that code ported between the dimensions reads the same;
  `Capsule3` holds it as a named `Segment` field rather than embedding it, so
  that `capsule.Height()` cannot be confused with the segment's `Length()`.

- **`Capsule3Box3` is EXACT**, and so is `Capsule3OBB3`. The approximation was
  not needed. Two branches, each closed-form:

  1. *The segment misses the box.* The squared distance from the box to a point
     travelling along the segment is a **convex, piecewise-quadratic** function of
     the segment parameter — convex because distance to a convex set is convex and
     the segment is affine in its parameter; piecewise-quadratic because within any
     stretch where each coordinate stays on the same side of its slab, the
     difference from the clamped point is affine. The pieces are delimited by the
     parameters at which a coordinate crosses a slab face: six at most. So the
     exact closest pair costs at most eight breakpoints and seven quadratics
     minimised in closed form, with no iteration and no tolerance.

     Given that pair at distance `d`, the MTV is `radius - d` along it. That is not
     obvious and is worth writing down: for `D` the Minkowski difference
     `box - segment` and `c` its point nearest the origin, moving by `lambda`
     along `-c/|c|` puts the origin beyond `D`'s supporting hyperplane at `c`, so
     the separation is at least `|c| + lambda`; and no shorter move can work,
     because any translation `t` reaches a separation of at most `|c| + |t|`.

  2. *The segment pierces the box.* The origin is then inside `D`, and its distance
     to a convex polytope's boundary is the least distance to a face plane — so a
     separating-axis minimum over `D`'s face normals is exact. `D` is a box swept
     along a segment, and a Minkowski sum's face normals are each operand's plus
     their edge-edge crosses; a segment contributes no faces and one edge, leaving
     the box's three axes and their three crosses with the segment direction. Six
     axes. Rounding by the radius adds the same amount to every one, so it does not
     change which is least.

  **The corner case the spec asked to have named** is therefore not a case where
  this is wrong — it is the case that rules out the shortcut. Running the
  sphere-box test at the point of the segment nearest the box's *centre* is right
  for a face contact and wrong on a corner, because the closest point to the centre
  is not the closest point to the box; a character on the corner of a step would
  get a normal pointing somewhere other than out of the corner and settle into it.
  `TestCapsule3Box3OnABoxCorner` is that case.

  The one place accuracy is soft is a segment grazing a corner at a gap of a few
  ulps. The depth stays accurate — it is a length — but the direction comes from
  subtracting nearly equal coordinates, and the returned normal can be some degrees
  out below about a float32 millionth. No threshold is applied to hide it: a pair
  that close is touching, the depth is right either way, and rounding the gap up to
  contact would trade a bounded direction error for an unbounded choice of which
  branch answers. Exact contact *is* exact, which is what the interval search's
  middle-of-the-interval return buys (see below).

- **`NewCapsule3Upright`'s height is the total, feet to crown**, caps included —
  the figure a designer means by "this character is 1.8 metres tall". A height
  below twice the radius has no answer and collapses to the sphere of that radius
  rather than erroring mid-frame.

## What landed

| Phase | Where |
|---|---|
| 1 | `linesegment3.go` (+`_test.go`) — the type and its distance routines; `capsule3.go` (+`_test.go`) — `Capsule3`, `NewCapsule3`, `NewCapsule3Upright`, `NewCapsule3FromSphere`, and the self-contained helpers |
| 2 | `collider/capsule3.go` — `Capsule3Sphere`, `Capsule3Capsule3` and both penetrations, plus `SphereCapsule3Penetration` |
| 3 | `collider/capsule3.go` — `Capsule3Box3`, `Capsule3OBB3`, their penetrations and the two reversed wrappers, over the shared `capsuleAlignedBoxPenetration` core |
| 4 | `collider/ray3.go` — `Ray3Capsule3`; 8 new rows in `penetration3d_consistency_test.go` (19 total) and 7 in `penetration3d_oracle_test.go` (16 total), all at `depthSlack: 1` |

The penetration matrix is now complete over all sixteen ordered pairs of
{Sphere, Box3, OBB3, Capsule3}.

### Two bugs the harnesses caught, worth knowing about

**A sign, in `LineSegment3.ClosestParameters`.** The degenerate branch projected
onto the other segment with the wrong sign, which clamped the parameter to an
endpoint whenever it should have landed in the interior. The unit table missed
it and the brute-force oracle in `linesegment3_test.go` found it immediately.

**An exact zero that was not one, in `closestOnSegmentToAlignedBox`.** Where the
segment passes through the box, the gap is constant at zero over an interval, and
the interval minimiser originally returned the interval's LOW end. That end is a
slab crossing computed by a division, so it evaluated to `1.2e-32` rather than to
zero — and `1.2e-32 > 0` sent the caller down its closest-pair branch, to
normalise a difference vector made entirely of rounding error. It returned a depth
of exactly the radius where the true answer was half again as much, because a small
box sits *inside* a fat capsule and has further to travel than a radius to get out.
Returning the interval's MIDDLE fixes it: an interior point evaluates to an exact
zero. `TestCapsule3Box3SegmentEndingInsideTheBox` pins it.

Only the 3D oracle caught the second one. The consistency harness did not, and
would not have: it is the same asymmetry PMP-570 recorded, where the harness's
X-axis rotations leave it blind to bugs the oracle sees. Confirmed again here by
injecting an inverse-rotation bug into `Capsule3OBB3Penetration` — the harness
stayed green, both oracle rows failed.

## Revision history

- v2 (2026-08-24) — implemented. Open questions resolved above: `LineSegment3`
  added, and the box pairs are exact rather than approximated.
- v1 (2026-08-14) — initial spec.
