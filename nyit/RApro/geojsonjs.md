Nice — here’s a compact, practical TypeScript guide with ready-to-use code for parsing and working with **GeoJSON** spatial data (Node or browser). It covers types, loading, common operations (iterating, bbox, filtering, point-in-polygon), using **@turf/turf** for spatial ops, and a simple spatial index example with **rbush**/**geojson-rbush** for fast spatial queries.

---

## Recommended packages

```bash
# Node / project
npm install @turf/turf @types/geojson geojson-rbush rbush
# or with yarn
# yarn add @turf/turf @types/geojson geojson-rbush rbush
```

* `@types/geojson` gives TypeScript types for GeoJSON objects (often installed as a dependency of other libs).
* `@turf/turf` provides spatial operations (bbox, booleanPointInPolygon, area, centroid, buffer, distance, etc.).
* `geojson-rbush` + `rbush` = spatial index for GeoJSON features.

---

## Basic types (use the GeoJSON types)

```ts
import type {
  GeoJsonObject,
  Feature,
  FeatureCollection,
  Geometry,
  Point,
} from "geojson";
```

---

## 1) Parse GeoJSON (string, file, or URL)

### Node: read from file

```ts
import fs from "fs/promises";
import type { FeatureCollection } from "geojson";

async function loadGeoJSONFromFile(path: string): Promise<FeatureCollection> {
  const raw = await fs.readFile(path, "utf8");
  const parsed = JSON.parse(raw) as GeoJsonObject;
  if (parsed.type !== "FeatureCollection") {
    throw new Error("Expected FeatureCollection");
  }
  return parsed as FeatureCollection;
}
```

### Browser or fetch

```ts
async function loadGeoJSONFromUrl(url: string): Promise<FeatureCollection> {
  const res = await fetch(url);
  const json = await res.json();
  if (json.type !== "FeatureCollection") throw new Error("Not a FeatureCollection");
  return json as FeatureCollection;
}
```

---

## 2) Iterate features and extract geometry/properties

```ts
import type { FeatureCollection, Feature, Geometry } from "geojson";

function listFeatures(fc: FeatureCollection) {
  for (const feature of fc.features as Feature[]) {
    const geom: Geometry | null = feature.geometry;
    const props = feature.properties;
    console.log("id:", feature.id, "type:", geom?.type, "props:", props);
  }
}
```

---

## 3) Compute bounding box & filter by bbox

Use Turf for bbox utilities.

```ts
import bbox from "@turf/bbox";
import booleanIntersects from "@turf/boolean-intersects";
import { bboxPolygon } from "@turf/turf";
import type { FeatureCollection } from "geojson";

function computeBBox(fc: FeatureCollection): number[] {
  // returns [minX, minY, maxX, maxY]
  return bbox(fc);
}

function filterWithinBBox(fc: FeatureCollection, queryBbox: number[]): FeatureCollection {
  const boxPoly = bboxPolygon(queryBbox);
  const filtered = fc.features.filter(f => booleanIntersects(f as any, boxPoly as any));
  return { type: "FeatureCollection", features: filtered };
}
```

---

## 4) Point in polygon and attribute filtering

```ts
import booleanPointInPolygon from "@turf/boolean-point-in-polygon";
import { point } from "@turf/helpers";
import type { FeatureCollection } from "geojson";

function featuresContainingPoint(fc: FeatureCollection, lon: number, lat: number) {
  const p = point([lon, lat]);
  return fc.features.filter(f => booleanPointInPolygon(p as any, f as any));
}

// property filter example
function filterByProperty(fc: FeatureCollection, key: string, value: unknown) {
  return {
    type: "FeatureCollection",
    features: fc.features.filter(f => (f.properties?.[key] ?? null) === value),
  };
}
```

---

## 5) Spatial index with geojson-rbush (fast bbox queries)

```ts
import GeoJSONRbush from "geojson-rbush";
import type { FeatureCollection, Feature } from "geojson";

function buildIndex(fc: FeatureCollection) {
  const idx = GeoJSONRbush();
  idx.load(fc);
  return idx;
}

// Query example: find features whose bbox intersects query bbox
function searchByBBox(idx: ReturnType<typeof GeoJSONRbush>, qbbox: [number, number, number, number]) {
  const [minX, minY, maxX, maxY] = qbbox;
  const results = idx.search({ type: "Feature", bbox: [minX, minY, maxX, maxY], geometry: null } as any);
  return results as Feature[];
}
```

Note: `geojson-rbush` expects GeoJSON features and accelerates searches for large datasets.

---

## 6) Examples of common spatial ops (turf)

```ts
import area from "@turf/area";
import centroid from "@turf/centroid";
import distance from "@turf/distance";
import { point } from "@turf/helpers";

function calcArea(feature: any) { return area(feature); }
function calcCentroid(feature: any) { return centroid(feature); }
function calcDistance(lon1:number, lat1:number, lon2:number, lat2:number) {
  return distance(point([lon1, lat1]), point([lon2, lat2]), { units: "kilometers" });
}
```

---

## 7) Example: full small script

A working Node example that loads a GeoJSON, builds an index, prints bbox, finds features containing a point and lists top 10 largest by area.

```ts
// example.ts
import fs from "fs/promises";
import GeoJSONRbush from "geojson-rbush";
import bbox from "@turf/bbox";
import booleanPointInPolygon from "@turf/boolean-point-in-polygon";
import { point } from "@turf/helpers";
import area from "@turf/area";
import type { FeatureCollection, Feature } from "geojson";

async function main() {
  const raw = await fs.readFile("./data.geojson", "utf8");
  const fc = JSON.parse(raw) as FeatureCollection;

  console.log("features:", fc.features.length, "type:", fc.type);

  const globalBbox = bbox(fc);
  console.log("global bbox:", globalBbox);

  // build index
  const idx = GeoJSONRbush();
  idx.load(fc);

  // point query
  const lon = -122.4, lat = 37.8; // SF
  const p = point([lon, lat]);
  const candidates = idx.search(p as any) as Feature[];
  const contains = candidates.filter(f => booleanPointInPolygon(p as any, f as any));
  console.log("features containing point:", contains.length);

  // top 10 by area
  const areas = fc.features.map(f => ({ f, a: area(f as any) })).sort((x,y) => y.a - x.a);
  console.log("Top 10 by area:");
  areas.slice(0,10).forEach((x,i) => console.log(i+1, x.a, x.f.id ?? x.f.properties?.name));
}

main().catch(err => console.error(err));
```

Compile/run:

```bash
tsc example.ts && node example.js
# or run with ts-node
npx ts-node example.ts
```

---

## 8) Notes about projections & coordinates

* GeoJSON coordinates are typically **WGS84 (EPSG:4326)** (lon, lat). Many turf functions assume this.
* If your data uses another projection (e.g., local meters), reproject using **proj4js** before spatial ops:

  ```ts
  import proj4 from "proj4";
  // reproject coordinates before building GeoJSON
  ```

---

## 9) Performance tips

* For large FeatureCollections (10k+ features), always **build a spatial index** (rbush / geojson-rbush) for bounding-box queries.
* Avoid iterating all features for every point query.
* Consider using vector tile libraries (`geojson-vt`) if serving map tiles.

---

## 10) TypeScript niceties

* Use `FeatureCollection<Geometry, Properties>` generics to have typed properties:

  ```ts
  import type { FeatureCollection, Feature } from "geojson";
  type Props = { id: number; name?: string };
  const fc: FeatureCollection = someJson as FeatureCollection;
  // or
  const fc2: FeatureCollection = JSON.parse(raw) as FeatureCollection;
  ```
* Add `@types/geojson` if needed:

  ```bash
  npm i --save-dev @types/geojson
  ```

---

If you tell me:

* whether you’re running **Node** or **browser**, and
* an example of the GeoJSON structure (Point-only, Polygons, MultiPolygons, feature properties),

I can produce a **tailored TypeScript module** (with function signatures and tests) to do exactly the operations you need (e.g., find nearest feature, cluster, snap, tile). Want that?

Excellent question — comparing **TypeScript vs Go** data types helps clarify how each language handles type systems. Let’s break it down carefully:

---

## 🧩 1. **Primitive Data Types in TypeScript**

TypeScript is a *superset of JavaScript*, so it shares JS’s type system (which is more abstracted and less memory-oriented than Go’s).

| TypeScript Type | Description                                                                                           | Example                                   | Go Equivalent                                                               |
| --------------- | ----------------------------------------------------------------------------------------------------- | ----------------------------------------- | --------------------------------------------------------------------------- |
| `number`        | Represents all numeric values (integer & floating-point). No distinction between `int`, `float`, etc. | `let x: number = 42;`                     | `int`, `float64`, `uint`, etc. (Go distinguishes sizes)                     |
| `bigint`        | For arbitrarily large integers (beyond 2⁵³-1). Introduced in ES2020.                                  | `let big: bigint = 123n;`                 | `big.Int` (from `math/big` package)                                         |
| `string`        | Sequence of characters.                                                                               | `let s: string = "hello";`                | `string`                                                                    |
| `boolean`       | True or false.                                                                                        | `let b: boolean = true;`                  | `bool`                                                                      |
| `symbol`        | Unique identifiers (used in object keys, etc.)                                                        | `let s = Symbol("id");`                   | No direct equivalent                                                        |
| `null`          | Represents a null value.                                                                              | `let n: null = null;`                     | `nil` (but Go uses nil for multiple types)                                  |
| `undefined`     | Variable declared but not assigned.                                                                   | `let u: undefined;`                       | No direct equivalent (`var x *T` with `nil` maybe)                          |
| `any`           | Opt-out of type checking.                                                                             | `let a: any = 10; a = "string";`          | `interface{}` (empty interface)                                             |
| `unknown`       | Like `any`, but type-safe — you must check before use.                                                | `let u: unknown = 10;`                    | `interface{}` with type assertions                                          |
| `void`          | No return value (e.g., functions that don’t return).                                                  | `function log(): void {}`                 | `func f() {}` (no return)                                                   |
| `never`         | Function never returns (e.g., throws or infinite loop).                                               | `function fail(): never { throw "err"; }` | No direct equivalent, closest: function that always panics or loops forever |

---

## 🧠 2. **Complex / Composite Types**

| TypeScript Type | Description               | Example                                    | Go Equivalent                                           |                                                            |
| --------------- | ------------------------- | ------------------------------------------ | ------------------------------------------------------- | ---------------------------------------------------------- |
| `array`         | Homogeneous collection    | `let arr: number[] = [1,2,3];`             | `[]int`                                                 |                                                            |
| `tuple`         | Fixed-length, typed array | `let t: [string, number] = ["age", 25];`   | Struct with fixed fields (e.g., `struct {string; int}`) |                                                            |
| `object`        | Key-value structure       | `let o: {name: string} = {name: "Alice"};` | `struct` or `map[string]any`                            |                                                            |
| `enum`          | Named constants           | `enum Color { Red, Green, Blue }`          | `const ( Red = iota; Green; Blue )`                     |                                                            |
| `union`         | Multiple possible types   | `let v: string                             | number;`                                                | No direct equivalent (can use `interface{}` + type switch) |
| `intersection`  | Combine multiple types    | `type A = X & Y;`                          | Go doesn’t have — closest: embedding structs            |                                                            |
| `interface`     | Type contract for objects | `interface Person { name: string }`        | Go interfaces (structural typing)                       |                                                            |
| `type` alias    | Custom type alias         | `type ID = string;`                        | `type ID string`                                        |                                                            |

---

## ⚖️ 3. **Numeric Comparison: Go vs TypeScript**

Go has **explicit numeric sizes**, but TypeScript does not.

| Go Type                               | TypeScript Equivalent | Notes                                         |
| ------------------------------------- | --------------------- | --------------------------------------------- |
| `int8`, `int16`, `int32`, `int64`     | `number`              | TypeScript doesn’t differentiate bit-widths.  |
| `uint8`, `uint16`, `uint32`, `uint64` | `number`              | No unsigned distinction.                      |
| `float32`, `float64`                  | `number`              | Only double precision (64-bit float in JS).   |
| `complex64`, `complex128`             | None                  | Need custom library or object representation. |

All JS/TS numbers are IEEE 754 **double-precision floating-point** (like Go’s `float64`).

---

## ⚙️ 4. **Example: Parsing GeoJSON Data in TypeScript**

Since you mentioned GeoJSON earlier, here’s a small example:

```typescript
type GeoJSONFeature = {
  type: "Feature";
  geometry: {
    type: "Point" | "Polygon" | "LineString";
    coordinates: number[] | number[][];
  };
  properties: Record<string, any>;
};

function parseGeoJSON(json: string): GeoJSONFeature[] {
  const data = JSON.parse(json);
  if (data.type === "FeatureCollection") {
    return data.features;
  } else if (data.type === "Feature") {
    return [data];
  } else {
    throw new Error("Invalid GeoJSON format");
  }
}

const sample = `{
  "type": "FeatureCollection",
  "features": [{
    "type": "Feature",
    "geometry": { "type": "Point", "coordinates": [102.0, 0.5] },
    "properties": { "prop0": "value0" }
  }]
}`;

const features = parseGeoJSON(sample);
console.log(features[0].geometry.coordinates);
```

---

✅ **Summary**

* TypeScript is *abstract* and unified (`number` covers all numeric types).
* Go is *low-level and explicit*, giving fine-grained control over memory and performance.
* You can think of TypeScript as *type-checked JavaScript* for web & app logic, and Go as a *compiled systems language* with strong typing and concurrency.

---

Would you like me to make a **side-by-side cheat sheet** showing TypeScript types vs Go types (with memory size and usage examples)?
