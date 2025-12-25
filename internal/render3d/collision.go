package render3d

import (
	"math"

	"github.com/mlange-42/ark/ecs"
)

// ColliderType3D identifies the type of 3D collider.
type ColliderType3D int

const (
	ColliderSphere ColliderType3D = iota
	ColliderBox
	ColliderCapsule
	ColliderMesh
)

// Collider3D represents a 3D collision shape.
type Collider3D struct {
	Type      ColliderType3D
	Offset    Vector3 // Offset from transform position
	Size      Vector3 // For box: dimensions. For sphere: X=radius. For capsule: X=radius, Y=height
	IsTrigger bool    // If true, only triggers events, no physics
	Layer     uint32  // Collision layer bitmask
	Mask      uint32  // Layers this collider checks against
}

// NewSphereCollider creates a sphere collider.
func NewSphereCollider(radius float32) Collider3D {
	return Collider3D{
		Type:  ColliderSphere,
		Size:  Vector3{X: radius},
		Layer: 1,
		Mask:  0xFFFFFFFF,
	}
}

// NewBoxCollider creates a box collider.
func NewBoxCollider(width, height, depth float32) Collider3D {
	return Collider3D{
		Type:  ColliderBox,
		Size:  Vector3{X: width, Y: height, Z: depth},
		Layer: 1,
		Mask:  0xFFFFFFFF,
	}
}

// NewCapsuleCollider creates a capsule collider.
func NewCapsuleCollider(radius, height float32) Collider3D {
	return Collider3D{
		Type:  ColliderCapsule,
		Size:  Vector3{X: radius, Y: height},
		Layer: 1,
		Mask:  0xFFFFFFFF,
	}
}

// Collision3D represents a collision event.
type Collision3D struct {
	EntityA   ecs.Entity
	EntityB   ecs.Entity
	Point     Vector3 // Contact point
	Normal    Vector3 // Collision normal
	Depth     float32 // Penetration depth
	IsTrigger bool
}

// CollisionSystem3D detects collisions between 3D entities.
type CollisionSystem3D struct {
	transforms  *ecs.Map1[Transform3D]
	colliders   *ecs.Map1[Collider3D]
	filter      *ecs.Filter2[Transform3D, Collider3D]
	collisions  []Collision3D
	OnCollision func(Collision3D)
}

// NewCollisionSystem3D creates a new 3D collision system.
func NewCollisionSystem3D() *CollisionSystem3D {
	return &CollisionSystem3D{
		collisions: make([]Collision3D, 0),
	}
}

// Init initializes the collision system.
func (s *CollisionSystem3D) Init(world *ecs.World) {
	s.transforms = ecs.NewMap1[Transform3D](world)
	s.colliders = ecs.NewMap1[Collider3D](world)
	s.filter = ecs.NewFilter2[Transform3D, Collider3D](world)
}

// Update checks for collisions.
func (s *CollisionSystem3D) Update(world *ecs.World, dt float32) {
	s.collisions = s.collisions[:0]

	// Collect all entities with colliders
	entities := make([]ecs.Entity, 0)

	query := s.filter.Query()
	for query.Next() {
		entities = append(entities, query.Entity())
	}

	// Check all pairs (O(n²) - can be optimized with spatial partitioning)
	for i := 0; i < len(entities); i++ {
		for j := i + 1; j < len(entities); j++ {
			entityA := entities[i]
			entityB := entities[j]

			transA := s.transforms.Get(entityA)
			transB := s.transforms.Get(entityB)
			colA := s.colliders.Get(entityA)
			colB := s.colliders.Get(entityB)

			// Check layer masks
			if colA.Layer&colB.Mask == 0 && colB.Layer&colA.Mask == 0 {
				continue
			}

			if collision, ok := s.checkCollision(entityA, entityB, transA, transB, colA, colB); ok {
				s.collisions = append(s.collisions, collision)
				if s.OnCollision != nil {
					s.OnCollision(collision)
				}
			}
		}
	}
}

// Draw is not needed for collision system.
func (s *CollisionSystem3D) Draw(world *ecs.World) {}

// GetCollisions returns the collisions from the last update.
func (s *CollisionSystem3D) GetCollisions() []Collision3D {
	return s.collisions
}

// checkCollision checks if two colliders are intersecting.
func (s *CollisionSystem3D) checkCollision(
	entityA, entityB ecs.Entity,
	transA, transB *Transform3D,
	colA, colB *Collider3D,
) (Collision3D, bool) {
	posA := transA.Position.Add(colA.Offset)
	posB := transB.Position.Add(colB.Offset)

	// Sphere-Sphere
	if colA.Type == ColliderSphere && colB.Type == ColliderSphere {
		return s.sphereSphere(
			entityA,
			entityB,
			posA,
			posB,
			colA.Size.X,
			colB.Size.X,
			colA.IsTrigger || colB.IsTrigger,
		)
	}

	// Box-Box (AABB)
	if colA.Type == ColliderBox && colB.Type == ColliderBox {
		return s.boxBox(entityA, entityB, posA, posB, colA.Size, colB.Size, colA.IsTrigger || colB.IsTrigger)
	}

	// Sphere-Box
	if colA.Type == ColliderSphere && colB.Type == ColliderBox {
		return s.sphereBox(
			entityA,
			entityB,
			posA,
			posB,
			colA.Size.X,
			colB.Size,
			colA.IsTrigger || colB.IsTrigger,
		)
	}

	if colA.Type == ColliderBox && colB.Type == ColliderSphere {
		col, ok := s.sphereBox(
			entityB,
			entityA,
			posB,
			posA,
			colB.Size.X,
			colA.Size,
			colA.IsTrigger || colB.IsTrigger,
		)
		if ok {
			// Swap entities back
			col.EntityA, col.EntityB = entityA, entityB
			col.Normal = col.Normal.Scale(-1)
		}

		return col, ok
	}

	return Collision3D{}, false
}

// sphereSphere checks sphere-sphere collision.
func (s *CollisionSystem3D) sphereSphere(
	entityA, entityB ecs.Entity,
	posA, posB Vector3,
	radiusA, radiusB float32,
	isTrigger bool,
) (Collision3D, bool) {
	diff := posB.Sub(posA)
	distSq := diff.Dot(diff)
	radiusSum := radiusA + radiusB

	if distSq > radiusSum*radiusSum {
		return Collision3D{}, false
	}

	dist := float32(math.Sqrt(float64(distSq)))

	normal := diff.Normalize()
	if dist == 0 {
		normal = Vector3{Y: 1}
	}

	return Collision3D{
		EntityA:   entityA,
		EntityB:   entityB,
		Point:     posA.Add(normal.Scale(radiusA)),
		Normal:    normal,
		Depth:     radiusSum - dist,
		IsTrigger: isTrigger,
	}, true
}

// boxBox checks AABB box-box collision.
func (s *CollisionSystem3D) boxBox(
	entityA, entityB ecs.Entity,
	posA, posB Vector3,
	sizeA, sizeB Vector3,
	isTrigger bool,
) (Collision3D, bool) {
	halfA := sizeA.Scale(0.5)
	halfB := sizeB.Scale(0.5)

	minA := posA.Sub(halfA)
	maxA := posA.Add(halfA)
	minB := posB.Sub(halfB)
	maxB := posB.Add(halfB)

	// Check overlap on all axes
	if maxA.X < minB.X || minA.X > maxB.X ||
		maxA.Y < minB.Y || minA.Y > maxB.Y ||
		maxA.Z < minB.Z || minA.Z > maxB.Z {
		return Collision3D{}, false
	}

	// Find penetration depth and normal
	overlapX := min32(maxA.X-minB.X, maxB.X-minA.X)
	overlapY := min32(maxA.Y-minB.Y, maxB.Y-minA.Y)
	overlapZ := min32(maxA.Z-minB.Z, maxB.Z-minA.Z)

	var (
		normal Vector3
		depth  float32
	)

	if overlapX < overlapY && overlapX < overlapZ {
		depth = overlapX

		if posA.X < posB.X {
			normal = Vector3{X: -1}
		} else {
			normal = Vector3{X: 1}
		}
	} else if overlapY < overlapZ {
		depth = overlapY

		if posA.Y < posB.Y {
			normal = Vector3{Y: -1}
		} else {
			normal = Vector3{Y: 1}
		}
	} else {
		depth = overlapZ

		if posA.Z < posB.Z {
			normal = Vector3{Z: -1}
		} else {
			normal = Vector3{Z: 1}
		}
	}

	return Collision3D{
		EntityA:   entityA,
		EntityB:   entityB,
		Point:     posA.Add(posB).Scale(0.5),
		Normal:    normal,
		Depth:     depth,
		IsTrigger: isTrigger,
	}, true
}

// sphereBox checks sphere-box collision.
func (s *CollisionSystem3D) sphereBox(
	entityA, entityB ecs.Entity,
	spherePos, boxPos Vector3,
	radius float32,
	boxSize Vector3,
	isTrigger bool,
) (Collision3D, bool) {
	halfSize := boxSize.Scale(0.5)

	// Find closest point on box to sphere center
	closest := Vector3{
		X: clamp32(spherePos.X, boxPos.X-halfSize.X, boxPos.X+halfSize.X),
		Y: clamp32(spherePos.Y, boxPos.Y-halfSize.Y, boxPos.Y+halfSize.Y),
		Z: clamp32(spherePos.Z, boxPos.Z-halfSize.Z, boxPos.Z+halfSize.Z),
	}

	diff := spherePos.Sub(closest)
	distSq := diff.Dot(diff)

	if distSq > radius*radius {
		return Collision3D{}, false
	}

	dist := float32(math.Sqrt(float64(distSq)))

	normal := diff.Normalize()
	if dist == 0 {
		normal = Vector3{Y: 1}
	}

	return Collision3D{
		EntityA:   entityA,
		EntityB:   entityB,
		Point:     closest,
		Normal:    normal,
		Depth:     radius - dist,
		IsTrigger: isTrigger,
	}, true
}

func min32(a, b float32) float32 {
	if a < b {
		return a
	}

	return b
}

func clamp32(value, minVal, maxVal float32) float32 {
	if value < minVal {
		return minVal
	}

	if value > maxVal {
		return maxVal
	}

	return value
}

// RaycastHit contains information about a raycast hit.
type RaycastHit struct {
	Entity   ecs.Entity
	Point    Vector3
	Normal   Vector3
	Distance float32
}

// Raycast3D performs a raycast and returns all hits.
func Raycast3D(
	world *ecs.World,
	origin, direction Vector3,
	maxDistance float32,
	layerMask uint32,
) []RaycastHit {
	hits := make([]RaycastHit, 0)

	transforms := ecs.NewMap1[Transform3D](world)
	colliders := ecs.NewMap1[Collider3D](world)
	filter := ecs.NewFilter2[Transform3D, Collider3D](world)

	query := filter.Query()
	for query.Next() {
		entity := query.Entity()
		transform := transforms.Get(entity)
		collider := colliders.Get(entity)

		if collider.Layer&layerMask == 0 {
			continue
		}

		pos := transform.Position.Add(collider.Offset)

		if collider.Type == ColliderSphere {
			if hit, ok := raySphere(origin, direction, pos, collider.Size.X, maxDistance); ok {
				hits = append(hits, RaycastHit{
					Entity:   entity,
					Point:    hit.point,
					Normal:   hit.normal,
					Distance: hit.distance,
				})
			}
		}
	}

	return hits
}

type rayHit struct {
	point    Vector3
	normal   Vector3
	distance float32
}

func raySphere(origin, direction, center Vector3, radius, maxDist float32) (rayHit, bool) {
	oc := origin.Sub(center)
	a := direction.Dot(direction)
	b := 2.0 * oc.Dot(direction)
	c := oc.Dot(oc) - radius*radius

	discriminant := b*b - 4*a*c
	if discriminant < 0 {
		return rayHit{}, false
	}

	t := (-b - float32(math.Sqrt(float64(discriminant)))) / (2.0 * a)
	if t < 0 || t > maxDist {
		return rayHit{}, false
	}

	point := origin.Add(direction.Scale(t))
	normal := point.Sub(center).Normalize()

	return rayHit{point: point, normal: normal, distance: t}, true
}
