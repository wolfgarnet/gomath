# gomath

Generic 2D/3D math and geometry library for Go games. Vectors, matrices, bounding volumes, collision, and spatial partitioning — all parameterized over a numeric type `T`.

## Quick Reference

```bash
go build ./...
go test ./...            # root + collider + spatial
go test ./collider/...   # one package
go test -run TestFoo ./  # one test
go test -bench=. ./spatial/...
```

## Project Structure

```
<root>/       Primitives (package gomath): Vector2, Vector3, Vector4, Matrix
                (col-major 3x3 *2D affine*), Matrix3 (col-major 3x3 *3D linear
                basis* — not interchangeable with Matrix), Matrix4 (col-major
                4x4 3D transform, GL upload order), AABB, Circle, Ellipsoid,
                LineSegment, Ray, Polygon, ConvexHull, plus scalar helpers
                (math.go, power.go, distance.go). Volume interfaces
                (CenteredVolume, SimpleVolume) in types.go.
collider/      Struct-based collision: bool intersection tests + penetration/MTV
                for every shape pair. (Was the old scalar `intersection` package.)
spatial/       Spatial partitioning: BoxedQuadtree, SimpleHashGrid, HashGrid,
                Quadtree, hash grids. (pixelmaster uses BoxedQuadtree + SimpleHashGrid.)
util/          Standalone helpers (distance).
test/         Cross-package tests. NOTE: test/vector_test.go still imports a
                deleted `gomath/g3n` package and does not build; the g3n/ directory
                was removed. Root, collider and spatial all build and pass.
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
- MTV/penetration normals point **shape1 → shape2** (the solver's convention). Keep every
  penetration function consistent with it; the consistency tests anchor this.
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
