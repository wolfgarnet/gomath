# gomath

Generic 2D/3D math and geometry library for Go games. Vectors, matrices, bounding volumes, collision, and spatial partitioning — all parameterized over a numeric type `T`.

## Quick Reference

```bash
go build ./...
go test ./...            # root + collider + spatial; all green
go test ./collider/...   # one package
go test -run TestFoo ./  # one test
go test -bench=. ./spatial/...
```

## Project Structure

```
<root>/       Primitives (package gomath): Vector2, Vector3, Vector4, Matrix
                (col-major 3x3 *2D affine*), Matrix3 (col-major 3x3 *3D linear
                basis* — not interchangeable with Matrix), Matrix4 (col-major
                4x4 3D transform, GL upload order), Quaternion, AABB, Circle,
                Ellipsoid, LineSegment, Ray, Polygon, ConvexHull, plus scalar
                helpers (math.go, power.go, distance.go). 3D: Sphere, Box3
                (axis-aligned), OBB3 (oriented), Capsule3, LineSegment3, Plane3,
                Ray3, Frustum. Volume interfaces (CenteredVolume, SimpleVolume)
                in types.go.
                NOTE: AABB, Circle, Ellipsoid, LineSegment, Ray and Polygon are
                all 2D despite some of the names — the 3D counterparts are Box3,
                Ellipsoid has none, LineSegment3, Ray3, and Polygon has none.
                There is no convex hull in 3D yet.
                Capsule3 is a LineSegment3 plus a radius, with no orientation
                field: the rotation is implicit in the endpoints. A degenerate
                segment makes it exactly a Sphere, and the tests lean on that
                reduction throughout — see docs/capsule3/spec.md.
collider/      Struct-based collision: bool intersection tests + penetration/MTV
                for every shape pair. (Was the old scalar `intersection` package.)
                The 3D penetration matrix is COMPLETE over
                {Sphere, Box3, OBB3, Capsule3}: every ordered pair has a
                `<A><B>Penetration` function, so callers never have to swap
                arguments and negate. Reversed forms (Box3Sphere, OBB3Sphere,
                OBB3Box3, SphereCapsule3, Box3Capsule3, OBB3Capsule3) are thin
                wrappers. The bool tests are NOT symmetric in the same way —
                SphereBox3, SphereOBB3 and the four Capsule3* tests exist but
                their reverses do not.
spatial/       Spatial partitioning: BoxedQuadtree, SimpleHashGrid, HashGrid,
                Quadtree, hash grids — all 2D. (pixelmaster uses BoxedQuadtree +
                SimpleHashGrid.) Grid3 (grid3.go) is the only 3D structure: a
                uniform grid with a per-axis cell size (zero on an axis leaves it
                unsubdivided), rebuild-per-tick, no toroidal wrap, offering
                QueryBox/QuerySphere, near-to-far RayCandidates and a Pairs
                broadphase sweep. See docs/spatial-3d/spec.md.
util/          Standalone helpers (distance).
                (test/ was removed: it existed only to benchmark gomath against a
                vendored copy of g3n's math32, which no longer exists.)
```

## Type System

- Everything is generic over `T Number` (`constraints.Signed | constraints.Float`) or
  `T Float` (float only). Defined in `vector.go`.
- **Integer `T` is storage-only for geometry.** Methods involving magnitude, angles, or
  trig (`Length`, `Normalize`, `Distance*`, `Angle*`, `Rotate*`, ...) compute in float64
  and truncate back to `T`, so results are meaningless for integer `T`. Free functions can
  be `[T Float]`; methods can't add a constraint beyond the receiver, so magnitude/trig
  methods live on `T Number` and truncate — use a float `T` when calling them.
- Squared-distance / cross-multiplied comparisons (`Contains`, `Intersects`, `ContainsCircle`)
  stay exact for integer `T`; area/circumference/`ClosestPoint` need float.

## Conventions

- Standard Go naming: PascalCase exported, camelCase unexported.
- Try to use real world variable names rather than single letters or obscure acronyms.
- Try to use single word struct receiver names rather than a single character.
- Try to use singular versions of slice names in plural rather than single characters. For example, element from elements.
- **Value receivers, no mutation — for small types.** Primitives are small (Vector2 = 16B,
  Vector3 = 24B, Vector4 = 32B) and pass in registers — value receivers benchmark ~2× faster
  than pointer, zero-alloc. Transforming a volume returns a **new** value via functional
  helpers (`Scaled`, `WithCenter`, `Translate`, `Translated/Rotated/Scaled` on Matrix). Do not
  add mutating setters or `*p` pointer variants to these.
- **Large array types (`Matrix4` = 128B, `Matrix3` = 72B at float64) are the exception.**
  The rule above was written when the library was 2D vectors and shapes; its rationale ("small,
  passes in registers") does not extend to a 128-byte array, which a value receiver copies in
  full on every call. Measured (pinned, medians): copying is free when arithmetic dominates —
  `Mul` is 35.17ns by value against 35.16ns by pointer — but costs real time when little work
  is done per byte read: `TransformPoint` over 4096 points is **1.54× faster** by pointer, and
  `GetPosition` over a slice of matrices **1.78× faster**, because a value receiver must touch
  every cache line of the matrix to read three floats. So for `Matrix3`/`Matrix4`:
  - **Pointer receiver** for methods that read the matrix to produce a *different* kind of
    value and are called in loops: `TransformPoint`, `TransformDirection`, `TransformVector4`,
    `TransformVector3`, `At`, `Column`, `GetPosition`, `GetScale`, `Basis`, `NormalMatrix`,
    `Floats32`. None of them mutate; the pointer only avoids the copy.
  - **Value receiver** for methods returning a new matrix or used in chained expressions:
    `Mul`, `MultiplyScalar`, `Transpose`, `Inverse`, `Determinant`, `WithPosition`,
    `Translated/Scaled/RotatedAxis`, `Equals`, `ApproxEquals`, `String`. Go requires
    addressability for a pointer-receiver call, so `m.Mul(b).Translated(...)` would otherwise
    stop compiling; `String` must stay a value receiver to keep `fmt.Stringer` on the value.

  The cost is that `m.Mul(b).TransformPoint(v)` no longer compiles and needs a local. That is
  the deliberate trade. When adding a method to a large matrix type, pick the receiver by which
  list it falls into, and benchmark **pinned** (`taskset -c 0 env GOMAXPROCS=1 ... -count=8`,
  report medians) — this machine's hybrid cores make unpinned results meaningless.
- **`SimpleVolume` is read-only** (center + half-extents) so both value and pointer forms
  satisfy it. Concrete transforms use the functional helpers above.
- Completeness additions go in a sibling `<type>_additions.go` (+ `_additions_test.go`),
  keeping the original file focused. Semantics/regression tests get their own `_semantics_test.go`
  / `_regression_test.go`.
- **Cross-shape collision lives in `collider/`, not on the root types** — the root types
  can't import `collider` (import cycle: collider → gomath). Root types carry only
  self-contained helpers (`PointAt`, `ClosestPoint`, `BoundingBox`, ...).
- **3D box overlap uses the scalar SAT form, not the obvious one.** `collider/obb3.go`
  precomputes the rotation of b in a's frame and tests fifteen axes in scalars. The direct
  transcription — building nine cross products as vectors — is both ~2.2x slower (measured,
  pinned) and *wrong* for near-parallel boxes, where the cross product is dominated by
  cancellation error. Exactly-parallel boxes are worse: every edge axis is zero, so whether
  the test says "separated" depends on whether the comparison is `>` or `>=`. If you touch
  that file, keep `TestOBB3ParallelBoxesOverlap` and `TestEdgeAxesAreNecessary` passing.
- **Capsule-vs-box is exact, and deliberately so.** `collider/capsule3.go` does not
  take the usual shortcut of running the sphere-box test at the segment's closest
  point to the box CENTRE, which is wrong on a corner — the case a character on a
  step hits. It branches instead: an exact segment-to-box distance (the squared
  distance is convex and piecewise quadratic in the segment parameter, so it is
  minimised in closed form over at most eight breakpoints) where the segment misses
  the box, and a six-axis SAT over the box's axes plus their crosses with the
  segment direction where it pierces. Both branches degenerate to the sphere
  functions when the segment does, which is what `TestCapsule3ReducesToSphereWhenDegenerate`
  checks. If you touch that file, keep `TestCapsule3Box3SegmentEndingInsideTheBox`
  passing — it pins why the interval search returns a MIDPOINT.
- MTV/penetration normals point **shape1 → shape2** (the solver's convention). Keep every
  penetration function consistent with it; the consistency tests anchor this —
  `collider/penetration_consistency_test.go` for 2D and
  `collider/penetration3d_consistency_test.go` for 3D. The 3D one is table-driven over
  every ordered pair and asserts four things: the normal points shape1 → shape2, moving
  shape2 by `penetration * normal` separates the pair, the swapped call negates the normal
  and keeps the depth, and a non-overlapping pair zeroes the other two returns. **Adding a
  3D pair means adding a row**, not writing a bespoke test — and a row in
  `collider/penetration3d_oracle_test.go` too. The two are not redundant: the
  harness fixes its rotations about X, which makes them invariant under
  transposing the rotation matrix, so it does NOT catch a normal mapped back to
  world space with the inverse rotation. The oracle does (verified by injecting
  exactly that bug into `Capsule3OBB3Penetration`: the harness stayed green).
- Real-world variable names over single letters; single-word receiver names; singular of a
  plural slice (`element` from `elements`). Repeat the type on every parameter
  (`f(a string, b string)`, not `f(a, b string)`). Avoid shadowing builtins (`min`/`max` →
  `lo`/`hi`).

## Testing

- `github.com/stretchr/testify/assert` for assertions; `require` when a failed check must
  stop the test (nil checks before deref, setup preconditions).
- Prefer **table-driven tests** — slice of case structs iterated with `t.Run(tc.name, ...)`.
  One-off tests are fine only when a table doesn't fit.
- Exact geometry (closest-point, penetration, ellipse distance) is cross-checked against a
  **brute-force oracle** and against reduction to a simpler shape (e.g. circular ellipse ==
  circle). Add these when touching exact-math paths.
- When integer truncation would make expected values ambiguous (odd-span `BoundingBox`,
  `Center`), write the test with float `T`.

## Primary Consumer

`../pixelmaster/` (an Ebiten game) is gomath's real consumer, wired via
`replace github.com/wolfgarnet/gomath => ../gomath` — changes land there directly, ~210
files import it. **Check pixelmaster call sites before making an API/design decision or
estimating blast radius**; real usage settles most questions. After a change, confirm
pixelmaster still builds (its only pre-existing failure is `TestReadFileSeparatorAgnostic`,
a Windows-path test on Linux — unrelated).

## Git

Main branch is **`master`** — branch from it, target it with PRs. There is no `main`.
Commit/push only when asked.
