# Capsule3 — Living Spec

**Linear:** [PMP-571](https://linear.app/pixel-master-pro/issue/PMP-571/capsule3-the-primitive-a-3d-character-collider-needs)

**Status:** not started.

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

- [ ] **Phase 1 — the type**: `Capsule3[T]` with `Segment gomath.LineSegment3[T]`
      (or two `Vector3` fields — see Open questions) and `Radius T`. Constructors
      for the upright character case (`NewCapsule3Upright(base, height, radius)`)
      and the general one. `BoundingBox() Box3[T]`, `Center()`, `ClosestPoint()`,
      the self-contained helpers the root types carry. **A segment-point distance
      helper is the whole engine of this feature** — if `distance.go` has no 3D
      segment-point routine, it gets one here.
- [ ] **Phase 2 — sphere and capsule pairs**: `Capsule3Sphere`,
      `Capsule3Capsule3`, and both penetrations. These are the cheap analytic
      ones: sphere is point-to-segment distance against `r1+r2`, capsule-capsule
      is segment-to-segment distance against the same. Land them first because
      they are exact and make good oracles for what follows.
- [ ] **Phase 3 — box pairs**: `Capsule3Box3`, `Capsule3OBB3` and their
      penetrations. Segment-vs-box is genuinely harder than the above and is
      where an approximation will be tempting. Whatever is chosen, say in this
      document what it is and where it is wrong — a capsule resting on a box
      corner is the case to name.
- [ ] **Phase 4 — `Ray3Capsule3`**, and the consistency + oracle rows for every
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

## Open questions

- Whether gomath needs a `LineSegment3[T]` type first. The root has
  `LineSegment` (2D, despite the name) and no 3D sibling. Two bare `Vector3`
  fields on the capsule avoids introducing a type nothing else asks for — but
  PMP-573's swept tests and PMP-575's wireframes both want to talk about 3D
  segments, so it may earn itself. Decide in Phase 1 and record the reasoning.
- Whether `Capsule3Box3` should be exact or an accepted approximation. Exactness
  costs a segment-vs-box distance routine; approximation risks a character
  sinking into a corner. Settle it in Phase 3 with the case written down.

## Revision history

- v1 (2026-08-14) — initial spec.
