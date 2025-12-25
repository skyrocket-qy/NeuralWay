package render3d

import (
	"math"
	"testing"
)

func TestVector3Add(t *testing.T) {
	a := Vector3{1, 2, 3}
	b := Vector3{4, 5, 6}
	result := a.Add(b)

	if result.X != 5 || result.Y != 7 || result.Z != 9 {
		t.Errorf("Add failed: got %v", result)
	}
}

func TestVector3Normalize(t *testing.T) {
	v := Vector3{3, 0, 4}
	n := v.Normalize()
	length := n.Length()

	if math.Abs(float64(length)-1.0) > 0.001 {
		t.Errorf("Normalize length should be 1, got %f", length)
	}
}

func TestOrbitCamera(t *testing.T) {
	cam := NewOrbitCamera(10, Vector3{})
	pos := cam.GetPosition()

	// Initial position should be at distance 10
	dist := pos.Length()
	if math.Abs(float64(dist)-10.0) > 0.1 {
		t.Errorf("Distance should be ~10, got %f", dist)
	}
}

func TestFPSCamera(t *testing.T) {
	cam := NewFPSCamera(Vector3{0, 5, 0})

	cam3d := cam.ToCamera3D()
	if cam3d.Position.Y != 5 {
		t.Errorf("Position Y should be 5, got %f", cam3d.Position.Y)
	}
}

func TestCreateCube(t *testing.T) {
	cube := CreateCube(2)

	if len(cube.Vertices) != 24 {
		t.Errorf("Cube should have 24 vertices, got %d", len(cube.Vertices))
	}

	if len(cube.Indices) != 36 {
		t.Errorf("Cube should have 36 indices, got %d", len(cube.Indices))
	}
}

func TestCreateSphere(t *testing.T) {
	sphere := CreateSphere(1, 16, 8)

	if len(sphere.Vertices) == 0 {
		t.Error("Sphere should have vertices")
	}
}

func TestMatrix4Identity(t *testing.T) {
	m := Identity()

	if m[0] != 1 || m[5] != 1 || m[10] != 1 || m[15] != 1 {
		t.Error("Identity diagonal should be 1")
	}
}

func TestSphereColliderCreation(t *testing.T) {
	col := NewSphereCollider(2.5)

	if col.Type != ColliderSphere {
		t.Error("Should be sphere type")
	}

	if col.Size.X != 2.5 {
		t.Errorf("Radius should be 2.5, got %f", col.Size.X)
	}
}

func TestBoxColliderCreation(t *testing.T) {
	col := NewBoxCollider(1, 2, 3)

	if col.Type != ColliderBox {
		t.Error("Should be box type")
	}

	if col.Size.X != 1 || col.Size.Y != 2 || col.Size.Z != 3 {
		t.Error("Box dimensions incorrect")
	}
}
