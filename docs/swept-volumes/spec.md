# Swept volumes — Living Spec

**Linear:** [PMP-573](https://linear.app/pixel-master-pro/issue/PMP-573/swept-volume-tests-so-a-fast-body-cannot-tunnel)

**Status:** not started.

## Goal

Add continuous collision tests to `collider/`: given a volume, a motion vector
for one step, and a static target, return the **time of impact** in `[0, 1]`
across that step. This is what stops a fast body passing clean through a thin
one between ticks.

## Depends on

- **[PMP-570](https://linear.app/pixel-master-pro/issue/PMP-570/complete-the-3d-penetration-matrix-and-a-consistency-harness-to-hold)** — *3D penetration matrix*.
  Read `docs/3d-penetration-matrix/spec.md` for the normal convention and the
  consistency harness. A swept test's contact normal obeys the same rule.
- **[PMP-571](https://linear.app/pixel-master-pro/issue/PMP-571/capsule3-the-primitive-a-3d-character-collider-needs)** — *Capsule3*.
  Read `docs/capsule3/spec.md`. **A swept sphere is a capsule** — the volume a
  sphere covers as it travels in a straight line is exactly the capsule whose
  segment is the path and whose radius is the sphere's. That identity is why
  most of this task is reuse rather than new mathematics, and it cannot be
  written before the type exists.

## Depended on by

- **PMP-581** (pixelmaster 3D queries) — hitscan and fast projectiles.

## The problem

Every test in `collider/` is **discrete**: it asks whether two volumes overlap
*right now*. Nothing asks whether they overlapped at any point between two
positions.

A body moving further than its own thickness in one tick is on one side before
the step and the other side after, and no discrete test placed at either end
will ever see it. At the engine's 60 TPS a body at 40 world-units/second covers
0.67 units per tick, so anything thinner than that is passable — and thin walls,
floors and projectiles are all ordinary content.

pixelmaster's 2D pipeline already dodges this once, for hitscan: instead of
testing the projectile's point position it tests a `LineSegment` spanning the
step (`simulation/systems/server/projectiles.go:73` → `math.IntersectionSegment`).
That is a swept *point*, done by hand at the call site.

In 3D, `Ray3Box3` / `Ray3OBB3` / `Ray3Sphere` already give the swept-point case.
What is missing is the swept **volume** — a moving body with a radius — which no
ray can express.

## Non-goals

- **Both volumes moving.** Only one side sweeps; the target is static within the
  step. Two moving bodies reduce to this by sweeping the relative motion, which
  is correct when neither rotates and is the standard simplification. Say so in
  the docs rather than silently assuming it.
- **Rotation during the sweep.** A body that rotates while translating does not
  sweep a shape any of this can express. Out of scope, and named so it is not
  mistaken for an oversight.
- Substepping, speculative contacts, or any solver policy. This task returns a
  number; what the caller does with it (advance to it, slide along the normal,
  re-sweep the remainder) is PMP-581's and the engine's business.
- The 2D pipeline. Untouched.

## Assumptions & constraints

- Everything in PMP-570's *Assumptions & constraints* applies.
- **Time is normalised to the step, not to seconds.** `t ∈ [0,1]` where 0 is the
  start position and 1 the end. The caller owns the timestep and does not have
  to tell this package about it.
- `t = 0` means *already overlapping at the start of the step*. That is a real
  and distinct answer from "no hit" and callers must be able to tell them apart —
  it means the discrete resolver should handle it, not the sweep.
- Generic over `T Float`; `float32` is the consumer's instantiation, and a swept
  test's discriminant is exactly the kind of arithmetic where `float32` precision
  shows up. Choose tolerances accordingly.

## Affected areas

- New: `collider/swept.go`, `collider/swept_test.go`.
- Possibly `collider/ray3.go`, if the swept-point cases are best expressed as
  thin wrappers over the existing ray casts.

## Phases

- [ ] **Phase 1 — swept sphere against sphere, Box3 and OBB3**: the core, and the
      cases the consumer needs first. Returns `(t T, normal Vector3[T], hit bool)`.
      Use the capsule identity where it simplifies: a swept sphere against a
      *static sphere* is a ray against a sphere of the summed radii, which is
      `Ray3Sphere` with an inflated target and needs no new mathematics at all.
      The box cases are the ones that need real work (Minkowski-inflated box with
      rounded edges; an exact solution is expensive and the usual approximation —
      ray against the inflated AABB — is wrong near corners, so name which one is
      implemented and where it errs).
- [ ] **Phase 2 — swept capsule**: against sphere, `Box3`, `OBB3` and capsule.
      Where Phase 1's approximations were accepted, the same ones apply; do not
      quietly upgrade one and not the other.
- [ ] **Phase 3 — the oracle**: for every swept pair, sample the motion at N
      substeps with the *discrete* test from PMP-570's matrix and check that the
      first substep reporting overlap brackets the analytic `t`. This is a
      genuinely strong oracle — it validates the continuous test against the
      discrete one that is already independently verified — and it is the reason
      PMP-570 is a dependency rather than merely a neighbour.
- [ ] **Phase 4 — the tunnelling regression test**: the case the whole task
      exists for. A body whose per-step motion exceeds a thin target's thickness,
      where the discrete test at both endpoints reports no overlap and the swept
      test reports a hit. If this test can be made to pass by a swept
      implementation that is subtly wrong, it is not specific enough.

## Design notes

**The identity that carries the feature.** The region a sphere of radius `r`
covers travelling from `a` to `b` is precisely `Capsule3{Segment: {a, b},
Radius: r}`. So "did this moving sphere touch anything" is "does this capsule
overlap anything" — a question PMP-571 already answers. That gives a *boolean*
for free. What it does **not** give is the time of impact, because the capsule
test says the paths crossed, not when. So:

- Use the capsule for the cheap **broad-phase reject**: no capsule overlap ⇒ no
  sweep, skip the expensive solve entirely.
- Solve for `t` only for the pairs that survive.

That two-stage shape is worth building in from the start; it is the difference
between a sweep costing something per candidate and per *hit*.

**Why the normal matters as much as the time.** A caller that advances to `t`
and stops has a body resting exactly on a surface, which the discrete test will
then report as touching-but-not-penetrating — the numerically worst place to be.
Callers will want to advance to `t` and then *slide* along the tangent, and
sliding needs the contact normal. A swept test returning only `t` will have to
grow the normal later; return it from the start.

## Open questions

- Whether swept-vs-`Box3` is exact or the inflated-AABB approximation. The
  approximation cuts corners — literally: it treats the rounded corner of the
  Minkowski sum as square, so a body clipping a box corner diagonally reports an
  impact slightly early. For a game that is often fine and occasionally visible.
  Decide in Phase 1 **with the error characterised**, not hand-waved.
- Whether `t = 0` (already overlapping) should be a distinct return or a separate
  `startedOverlapping bool`. Folding it into `t` risks a caller treating "already
  stuck" as "hits immediately" and advancing zero distance forever.

## Revision history

- v1 (2026-08-14) — initial spec.
