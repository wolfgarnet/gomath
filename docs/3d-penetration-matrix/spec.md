# 3D penetration matrix — Living Spec

**Linear:** [PMP-570](https://linear.app/pixel-master-pro/issue/PMP-570/complete-the-3d-penetration-matrix-and-a-consistency-harness-to-hold)

**Status:** not started.

> This is the first spec in gomath, so it also establishes the convention:
> `docs/<topic-slug>/spec.md`, one directory per spec'd task, phases checked off
> in place, revision history at the bottom. It follows the template in
> `../pixelmaster/CLAUDE.md`, which is where the sibling repos' convention comes
> from. gorendr3d has a richer variant; gomath does not need it yet.

## Goal

Close the one missing function in gomath's 3D penetration set and put a
consistency harness under all of it, so that the 3D collision pipeline being
built in pixelmaster (PMP-576 … PMP-581) dispatches into a matrix that is
complete and whose normals all point the same way.

This is the root of the whole 3D-collision dependency graph. Nothing else in it
can be trusted before this lands, because every downstream task inherits these
functions' sign convention.

## Depends on

Nothing. This is the root task.

## Depended on by

- **PMP-571** (Capsule3) — the capsule's penetration functions slot into the
  harness this builds rather than re-deriving the convention.
- **PMP-573** (swept volumes) — same.
- **PMP-577** (pixelmaster's 3D narrow phase) — dispatches into this matrix.

## What exists today

`collider/` holds these 3D functions, written across three commits
(`7525e97`, `4de3177`, and the OBB work before them):

| File | Functions |
|---|---|
| `collider/sphere3.go` | `SphereSphere`, `SphereSpherePenetration`, `SphereBox3`, `SphereBox3Penetration`, `SphereOBB3` |
| `collider/box3.go` | `Box3Box3`, `Box3Box3Penetration`, `Box3OBB3`, `OBB3Box3`, `Box3OBB3Penetration` |
| `collider/obb3.go` | `OBB3OBB3`, `OBB3OBB3Penetration` |
| `collider/ray3.go` | `Ray3Plane3`, `Ray3Sphere`, `Ray3Box3`, `Ray3OBB3` |

**`SphereOBB3Penetration` does not exist.** `SphereOBB3` returns a bool and
there is no MTV counterpart. Every other 3D pair has one.

A sphere against an oriented box is the pair a character-vs-scenery test
produces most often, so this hole sits directly in the path of the work that
depends on it.

## Non-goals

- Capsules. They are PMP-571 and land on top of this.
- Continuous / swept tests. They are PMP-573.
- Convex hulls, triangle meshes, or anything requiring GJK/EPA. The 3D set stays
  analytic primitives. A mesh gets a proxy volume instead (PMP-574).
- Touching the 2D functions. They are in use by the shipped game and this task
  has no reason to go near them.

## Assumptions & constraints

- **The MTV convention is already law and is not up for renegotiation here.**
  gomath's CLAUDE.md states it: *"MTV/penetration normals point shape1 → shape2
  (the solver's convention)."* The 2D `collider/penetration_consistency_test.go`
  anchors it. This task extends that anchoring to 3D; it does not change the
  rule.
- **`float32` is the consumer's type.** pixelmaster uses `float32` throughout.
  The functions stay generic over `T Float`, but the tolerances the tests pick
  must be ones `float32` can actually hold — a `1e-9` epsilon written against
  `float64` intuition will fail the consumer's instantiation.
- Value receivers, no mutation, per the repo convention for small types.
- **`collider` may import `gomath`, never the reverse.** Cross-shape collision
  cannot live on the root types (import cycle).
- Benchmarks, if any are added, must be **pinned**
  (`taskset -c 0 env GOMAXPROCS=1 ... -count=8`, medians reported). This
  machine's hybrid cores make unpinned numbers meaningless — the repo's CLAUDE.md
  says so and §8.36 in gorendr3d has a worked example of it going wrong (p99s of
  10.2 ms and 77.8 ms from two near-identical runs).

## Affected areas

- `collider/sphere3.go` — the new function.
- `collider/penetration_consistency_test.go` — or a 3D sibling; the phase decides.
- New: a 3D brute-force oracle test file.
- `CLAUDE.md` — its Project Structure section is **stale** and says *"there is no
  Sphere, Ray3, Plane or Frustum yet"*, which stopped being true three commits
  ago. Fix it in this task; a downstream agent reading it will otherwise conclude
  the 3D primitives do not exist.

## Phases

- [ ] **Phase 1 — `SphereOBB3Penetration`**: the missing MTV. Transform the
      sphere's centre into the box's frame, clamp to the box's half-extents to
      find the closest point, and map back — the same shape as
      `SphereBox3Penetration` with a rotation on either side. Handle the
      centre-inside-the-box case, where the closest point is the centre itself
      and the correct axis is the least-penetrated face, not a zero-length
      vector.
- [ ] **Phase 2 — the 3D consistency harness**: for every 3D penetration pair,
      assert (a) the normal points shape1 → shape2, (b) translating shape2 by
      `penetration * normal` separates them to within tolerance, (c) the swapped
      call returns the negated normal and the same magnitude, and (d) a
      non-overlapping pair returns `overlapping == false` and leaves the other
      returns at zero. Table-driven, one row per pair.
- [ ] **Phase 3 — a brute-force oracle for the exact-math pairs**: sample points
      on a grid, take the true minimum separation, and check the analytic MTV
      against it. The repo already does this for 2D ellipse distance
      (`collider/ellipsoid_oracle_test.go`) — same technique, 3D shapes.
- [ ] **Phase 4 — correct `CLAUDE.md`**: the Project Structure paragraph, and a
      line stating which 3D pairs have penetration functions so the next reader
      does not have to grep for it.

## Design notes

**The interface downstream tasks should code against.** Every 3D penetration
function has this shape, and PMP-577's dispatch table is written from it:

```go
func <Shape1><Shape2>Penetration[T gomath.Float](
    a gomath.<Shape1>[T],
    b gomath.<Shape2>[T],
) (penetration T, normal gomath.Vector3[T], overlapping bool)
```

- `overlapping == false` ⇒ the other two returns are meaningless; callers must
  not read them. (The 2D code has call sites that do; do not copy that.)
- `normal` is unit length and points **from `a` toward `b`**.
- `penetration` is the depth along `normal`, non-negative.

**Why the sign convention keeps biting.** The consumer wants the normal pointing
the other way about half the time, so `pixelmaster/math/penetration.go` has
`normal = normal.MultiplyScalar(-1)` scattered at individual call sites — see
its `PenetrationCircle` ellipsoid branch, `PenetrationAABB` circle branch, and
four more. Each one is a place a future edit can silently invert and push bodies
*together*. gomath's job is to make the convention uniform enough that the
consumer can flip once, centrally. Phase 2's part (c) is what proves it is
uniform.

**The centre-inside case is where these functions are usually wrong.** The
closest-point-on-box formulation degenerates when the sphere centre is inside
the box: the closest point equals the centre, the difference vector is zero, and
a naive `Normalize()` gives NaN or `(0,0,0)` depending on the implementation.
The right answer is the axis of least penetration among the six faces. This is
worth a named test, because a body deeply overlapped is exactly the state the
resolver has to dig out of and exactly the state that is rare in casual testing.

## Open questions

- Whether the 3D consistency table lives in the existing
  `penetration_consistency_test.go` or a `penetration3d_consistency_test.go`.
  Leaning separate: the 2D table is already long and the two share no fixtures.
- Whether `SphereOBB3Penetration` should be joined by an `OBB3SpherePenetration`
  wrapper for symmetry (`Box3OBB3`/`OBB3Box3` set that precedent for the bool
  tests). Cheap to add, and it removes an argument-order trap from the consumer.

## Revision history

- v1 (2026-08-14) — initial spec.
