package render3d

import "math"

// CreateCube generates a cube mesh centered at origin.
func CreateCube(size float32) Mesh3D {
	half := size / 2
	vertices := []Vector3{
		// Front face
		{-half, -half, half},
		{half, -half, half},
		{half, half, half},
		{-half, half, half},
		// Back face
		{-half, -half, -half},
		{-half, half, -half},
		{half, half, -half},
		{half, -half, -half},
		// Top face
		{-half, half, -half},
		{-half, half, half},
		{half, half, half},
		{half, half, -half},
		// Bottom face
		{-half, -half, -half},
		{half, -half, -half},
		{half, -half, half},
		{-half, -half, half},
		// Right face
		{half, -half, -half},
		{half, half, -half},
		{half, half, half},
		{half, -half, half},
		// Left face
		{-half, -half, -half},
		{-half, -half, half},
		{-half, half, half},
		{-half, half, -half},
	}

	indices := []uint32{
		0, 1, 2, 0, 2, 3, // Front
		4, 5, 6, 4, 6, 7, // Back
		8, 9, 10, 8, 10, 11, // Top
		12, 13, 14, 12, 14, 15, // Bottom
		16, 17, 18, 16, 18, 19, // Right
		20, 21, 22, 20, 22, 23, // Left
	}

	return Mesh3D{
		Vertices: vertices,
		Indices:  indices,
		Visible:  true,
	}
}

// CreateSphere generates a UV sphere mesh.
func CreateSphere(radius float32, segments, rings int) Mesh3D {
	vertices := make([]Vector3, 0)
	indices := make([]uint32, 0)

	for i := 0; i <= rings; i++ {
		phi := math.Pi * float64(i) / float64(rings)
		for j := 0; j <= segments; j++ {
			theta := 2 * math.Pi * float64(j) / float64(segments)

			x := float32(math.Sin(phi) * math.Cos(theta))
			y := float32(math.Cos(phi))
			z := float32(math.Sin(phi) * math.Sin(theta))

			vertices = append(vertices, Vector3{
				X: x * radius,
				Y: y * radius,
				Z: z * radius,
			})
		}
	}

	// Generate indices
	for i := range rings {
		for j := range segments {
			first := uint32(i*(segments+1) + j)
			second := first + uint32(segments) + 1

			indices = append(indices, first, second, first+1)
			indices = append(indices, second, second+1, first+1)
		}
	}

	return Mesh3D{
		Vertices: vertices,
		Indices:  indices,
		Visible:  true,
	}
}

// CreatePlane generates a horizontal plane mesh.
func CreatePlane(width, depth float32, subdivisionsX, subdivisionsZ int) Mesh3D {
	vertices := make([]Vector3, 0)
	indices := make([]uint32, 0)

	halfW := width / 2
	halfD := depth / 2
	stepX := width / float32(subdivisionsX)
	stepZ := depth / float32(subdivisionsZ)

	for z := 0; z <= subdivisionsZ; z++ {
		for x := 0; x <= subdivisionsX; x++ {
			vertices = append(vertices, Vector3{
				X: -halfW + float32(x)*stepX,
				Y: 0,
				Z: -halfD + float32(z)*stepZ,
			})
		}
	}

	for z := range subdivisionsZ {
		for x := range subdivisionsX {
			topLeft := uint32(z*(subdivisionsX+1) + x)
			topRight := topLeft + 1
			bottomLeft := topLeft + uint32(subdivisionsX+1)
			bottomRight := bottomLeft + 1

			indices = append(indices, topLeft, bottomLeft, topRight)
			indices = append(indices, topRight, bottomLeft, bottomRight)
		}
	}

	return Mesh3D{
		Vertices: vertices,
		Indices:  indices,
		Visible:  true,
	}
}

// CreateCylinder generates a cylinder mesh.
func CreateCylinder(radius, height float32, segments int) Mesh3D {
	vertices := make([]Vector3, 0)
	indices := make([]uint32, 0)

	halfH := height / 2

	// Top and bottom center vertices
	topCenter := uint32(0)

	vertices = append(vertices, Vector3{Y: halfH})
	bottomCenter := uint32(1)

	vertices = append(vertices, Vector3{Y: -halfH})

	// Ring vertices
	for i := 0; i <= segments; i++ {
		theta := 2 * math.Pi * float64(i) / float64(segments)
		x := float32(math.Cos(theta)) * radius
		z := float32(math.Sin(theta)) * radius

		// Top ring
		vertices = append(vertices, Vector3{X: x, Y: halfH, Z: z})
		// Bottom ring
		vertices = append(vertices, Vector3{X: x, Y: -halfH, Z: z})
	}

	// Generate indices for caps and sides
	for i := range segments {
		topA := uint32(2 + i*2)
		topB := uint32(2 + (i+1)*2)
		bottomA := topA + 1
		bottomB := topB + 1

		// Top cap
		indices = append(indices, topCenter, topA, topB)
		// Bottom cap
		indices = append(indices, bottomCenter, bottomB, bottomA)
		// Side
		indices = append(indices, topA, bottomA, topB)
		indices = append(indices, topB, bottomA, bottomB)
	}

	return Mesh3D{
		Vertices: vertices,
		Indices:  indices,
		Visible:  true,
	}
}
