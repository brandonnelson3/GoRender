package gfx

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/brandonnelson3/GoRender/loader"
)

type material struct {
	name    string
	diffuse uint32
}

type group struct {
	start, end int32

	mat material
}

// Object is a wrapper around the contents of a model file.
type Object struct {
	groups   []group
	vertices []Vertex
}

const (
	// Obj
	commentPrefix = "#"
	mtllibPrefix  = "mtllib "
	mtllibFormat  = mtllibPrefix + "%s"
	gPrefix       = "g "
	usemtlPrefix  = "usemtl "
	usemtlFormat  = usemtlPrefix + "%s"
	vPrefix       = "v "
	vFormat       = vPrefix + "%f %f %f"
	vnPrefix      = "vn "
	vnFormat      = vnPrefix + "%f %f %f"
	vtPrefix      = "vt "
	vtFormat      = vtPrefix + "%f %f"
	fPrefix       = "f "
	fFormat       = fPrefix + "%d/%d/%d %d/%d/%d %d/%d/%d"

	// Mtl
	newMtlPrefix = "newmtl "
	newMtlFormat = newMtlPrefix + "%s"
	mapKdPrefix  = "map_Kd "
	mapKdFormat  = mapKdPrefix + "%s"
)

// GetChunkedRenderable builds the renderable for this Object.
func (o *Object) GetChunkedRenderable() *VAORenderable {
	portions := []RenderablePortion{}
	for _, g := range o.groups {
		portions = append(portions, RenderablePortion{g.start, g.end - g.start, g.mat.diffuse})
	}
	return NewChunkedRenderable(o.vertices, portions)
}

// LoadObjFile loads the provided .obj file.
func LoadObjFile(file string) (*Object, error) {
	r, err := loader.Load(file)
	if err != nil {
		return nil, err
	}
	result := Object{}
	scanner := bufio.NewScanner(r)

	var verts []mgl32.Vec3
	var normals []mgl32.Vec3
	var texCoords []mgl32.Vec2
	mtllib := make(map[string]material)
	var g group

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, commentPrefix) {
			continue
		}
		if strings.HasPrefix(line, mtllibPrefix) {
			mtllibFilename := ""
			if _, err := fmt.Sscanf(line, mtllibFormat, &mtllibFilename); err != nil {
				return nil, fmt.Errorf("Got error while parsing mtllib line: %s, %v", line, err)
			}
			if len(mtllibFilename) == 0 {
				return nil, fmt.Errorf("Got empty filename for mtllib in: %s", file)
			}
			newmtllib, err := loadMtlFile(mtllibFilename)
			if err != nil {
				return nil, fmt.Errorf("Got error while parsing mtl file %s: %v", mtllibFilename, err)
			}
			// Merge the new map with any previously seen maps.
			for k, v := range newmtllib {
				mtllib[k] = v
			}
		}
		if strings.HasPrefix(line, gPrefix) {
			if g.end > g.start {
				result.groups = append(result.groups, g)
				g = group{
					start: int32(len(result.vertices)) + 1,
					end:   int32(len(result.vertices)) + 1,
				}
			}
		}
		if strings.HasPrefix(line, usemtlPrefix) {
			usemtlValue := ""
			if _, err := fmt.Sscanf(line, usemtlFormat, &usemtlValue); err != nil {
				return nil, fmt.Errorf("Got error while parsing usemtl line: %s, %v", line, err)
			}
			if len(usemtlValue) == 0 {
				return nil, fmt.Errorf("Got empty material name for usemtl in: %s", file)
			}
			mat, ok := mtllib[usemtlValue]
			if !ok {
				return nil, fmt.Errorf("Attempted to use material that has not been loaded: %s", usemtlValue)
			}
			g.mat = mat
		}
		if strings.HasPrefix(line, vPrefix) {
			vert, err := parseVec3Line(vFormat, line)
			if err != nil {
				return nil, fmt.Errorf("Got error while parsing vertex line: %s, %v", line, err)
			}
			verts = append(verts, vert)
		}
		if strings.HasPrefix(line, vnPrefix) {
			normal, err := parseVec3Line(vnFormat, line)
			if err != nil {
				return nil, fmt.Errorf("Got error while parsing normal line: %s, %v", line, err)
			}
			normals = append(normals, normal)
		}
		if strings.HasPrefix(line, vtPrefix) {
			texCoord, err := parseVec2Line(vtFormat, line)
			if err != nil {
				return nil, fmt.Errorf("Got error while parsing texcoord line: %s, %v", line, err)
			}
			texCoords = append(texCoords, texCoord)
		}
		if strings.HasPrefix(line, fPrefix) {
			v1, t1, n1, v2, t2, n2, v3, t3, n3, err := parseFaceLine(fFormat, line)
			if err != nil {
				return nil, fmt.Errorf("Got error while parsing face line: %s, %v", line, err)
			}
			vert1 := Vertex{
				Vert: verts[v1-1],
				Norm: normals[n1-1],
				UV:   texCoords[t1-1],
			}
			vert2 := Vertex{
				Vert: verts[v2-1],
				Norm: normals[n2-1],
				UV:   texCoords[t2-1],
			}
			vert3 := Vertex{
				Vert: verts[v3-1],
				Norm: normals[n3-1],
				UV:   texCoords[t3-1],
			}
			g.end += 3
			result.vertices = append(result.vertices, vert1, vert2, vert3)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if g.end > g.start {
		result.groups = append(result.groups, g)
		g = group{
			start: int32(len(result.vertices)),
		}
	}
	return &result, nil
}

func parseVec3Line(format, line string) (mgl32.Vec3, error) {
	result := mgl32.Vec3{}
	if _, err := fmt.Sscanf(line, format, &result[0], &result[1], &result[2]); err != nil {
		return result, err
	}
	return result, nil
}

func parseVec2Line(format, line string) (mgl32.Vec2, error) {
	result := mgl32.Vec2{}
	if _, err := fmt.Sscanf(line, format, &result[0], &result[1]); err != nil {
		return result, err
	}
	return result, nil
}

func parseFaceLine(format, line string) (uint32, uint32, uint32, uint32, uint32, uint32, uint32, uint32, uint32, error) {
	var v1, n1, t1, v2, n2, t2, v3, n3, t3 uint32
	if _, err := fmt.Sscanf(line, format, &v1, &t1, &n1, &v2, &t2, &n2, &v3, &t3, &n3); err != nil {
		return v1, t1, n1, v2, t2, n2, v3, t3, n3, err
	}
	return v1, t1, n1, v2, t2, n2, v3, t3, n3, nil
}

func loadMtlFile(file string) (map[string]material, error) {
	r, err := loader.Load(file)
	if err != nil {
		return nil, err
	}
	result := make(map[string]material)
	scanner := bufio.NewScanner(r)

	var mat *material
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, newMtlPrefix) {
			materialName := ""
			if _, err := fmt.Sscanf(line, newMtlFormat, &materialName); err != nil {
				return nil, fmt.Errorf("Got error while parsing newmtl line: %s, %v", line, err)
			}
			if len(materialName) == 0 {
				return nil, fmt.Errorf("Got empty name for material in: %s", file)
			}
			if mat != nil {
				result[mat.name] = *mat
			}
			mat = &material{
				name: materialName,
			}
		}
		if strings.HasPrefix(line, mapKdPrefix) {
			diffuseFile := ""
			if _, err := fmt.Sscanf(line, mapKdFormat, &diffuseFile); err != nil {
				return nil, fmt.Errorf("Got error while parsing mapKd line: %s, %v", line, err)
			}
			if len(diffuseFile) == 0 {
				return nil, fmt.Errorf("Got empty filename for mapKd in: %s", file)
			}
			diffuseTexture, err := LoadTexture(diffuseFile)
			if err != nil {
				return nil, fmt.Errorf("Error while loading diffuse texture for material: %s, from file: %s", diffuseFile, file)
			}
			mat.diffuse = diffuseTexture
		}
	}
	if mat != nil {
		result[mat.name] = *mat
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
