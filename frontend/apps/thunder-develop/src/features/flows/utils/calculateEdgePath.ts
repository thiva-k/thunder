/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import {type Node, Position} from '@xyflow/react';

// ============================================================================
// Types
// ============================================================================

interface Point {
  x: number;
  y: number;
}

interface Rectangle {
  left: number;
  right: number;
  top: number;
  bottom: number;
}

export interface EdgePathResult {
  path: string;
  centerX: number;
  centerY: number;
}

export interface EdgeInput {
  id: string;
  sourceX: number;
  sourceY: number;
  targetX: number;
  targetY: number;
  sourcePosition: Position;
  targetPosition: Position;
}

interface Segment {
  edgeId: string;
  segmentIndex: number;
  type: 'horizontal' | 'vertical';
  p1: Point;
  p2: Point;
  // Normalized coordinates for comparison
  fixedCoord: number; // Y for horizontal, X for vertical
  minVar: number; // min X for horizontal, min Y for vertical
  maxVar: number; // max X for horizontal, max Y for vertical
}

// ============================================================================
// Configuration
// ============================================================================

const NODE_MARGIN = 10;
const EXIT_PADDING = 15;
const SEPARATION_DISTANCE = 12; // Distance between parallel edges
const OVERLAP_THRESHOLD = 3; // Segments within this distance are considered overlapping

// ============================================================================
// Geometry Utilities
// ============================================================================

function getNodeBounds(node: Node, margin: number): Rectangle {
  const width = node.measured?.width ?? node.width ?? 150;
  const height = node.measured?.height ?? node.height ?? 50;

  return {
    left: node.position.x - margin,
    right: node.position.x + width + margin,
    top: node.position.y - margin,
    bottom: node.position.y + height + margin,
  };
}

function pointInRect(x: number, y: number, rect: Rectangle): boolean {
  return x >= rect.left && x <= rect.right && y >= rect.top && y <= rect.bottom;
}

function hSegmentIntersectsRect(x1: number, x2: number, y: number, rect: Rectangle): boolean {
  if (y < rect.top || y > rect.bottom) return false;
  const minX = Math.min(x1, x2);
  const maxX = Math.max(x1, x2);
  return maxX > rect.left && minX < rect.right;
}

function vSegmentIntersectsRect(x: number, y1: number, y2: number, rect: Rectangle): boolean {
  if (x < rect.left || x > rect.right) return false;
  const minY = Math.min(y1, y2);
  const maxY = Math.max(y1, y2);
  return maxY > rect.top && minY < rect.bottom;
}

function hBlocked(x1: number, x2: number, y: number, obstacles: Rectangle[]): boolean {
  return obstacles.some((r) => hSegmentIntersectsRect(x1, x2, y, r));
}

function vBlocked(x: number, y1: number, y2: number, obstacles: Rectangle[]): boolean {
  return obstacles.some((r) => vSegmentIntersectsRect(x, y1, y2, r));
}

// ============================================================================
// Exit Point Calculation
// ============================================================================

function getExitPoint(handleX: number, handleY: number, position: Position, obstacles: Rectangle[]): Point {
  const containers = obstacles.filter((r) => pointInRect(handleX, handleY, r));

  if (containers.length === 0) {
    switch (position) {
      case Position.Right:
        return {x: handleX + EXIT_PADDING, y: handleY};
      case Position.Left:
        return {x: handleX - EXIT_PADDING, y: handleY};
      case Position.Bottom:
        return {x: handleX, y: handleY + EXIT_PADDING};
      case Position.Top:
        return {x: handleX, y: handleY - EXIT_PADDING};
      default:
        return {x: handleX, y: handleY};
    }
  }

  switch (position) {
    case Position.Right: {
      let exitX = handleX;
      containers.forEach((r) => {
        exitX = Math.max(exitX, r.right);
      });
      return {x: exitX + EXIT_PADDING, y: handleY};
    }
    case Position.Left: {
      let exitX = handleX;
      containers.forEach((r) => {
        exitX = Math.min(exitX, r.left);
      });
      return {x: exitX - EXIT_PADDING, y: handleY};
    }
    case Position.Bottom: {
      let exitY = handleY;
      containers.forEach((r) => {
        exitY = Math.max(exitY, r.bottom);
      });
      return {x: handleX, y: exitY + EXIT_PADDING};
    }
    case Position.Top: {
      let exitY = handleY;
      containers.forEach((r) => {
        exitY = Math.min(exitY, r.top);
      });
      return {x: handleX, y: exitY - EXIT_PADDING};
    }
    default:
      return {x: handleX, y: handleY};
  }
}

// ============================================================================
// Path Utilities
// ============================================================================

function isPathClear(points: Point[], obstacles: Rectangle[]): boolean {
  for (let i = 0; i < points.length - 1; i += 1) {
    const p1 = points[i];
    const p2 = points[i + 1];

    const isHorizontal = Math.abs(p1.y - p2.y) < 1;
    const isVertical = Math.abs(p1.x - p2.x) < 1;

    if (isHorizontal && hBlocked(p1.x, p2.x, p1.y, obstacles)) return false;
    if (isVertical && vBlocked(p1.x, p1.y, p2.y, obstacles)) return false;
  }
  return true;
}

function pathLength(points: Point[]): number {
  let len = 0;
  for (let i = 1; i < points.length; i += 1) {
    len += Math.abs(points[i].x - points[i - 1].x) + Math.abs(points[i].y - points[i - 1].y);
  }
  return len;
}

// ============================================================================
// Corridor Finding
// ============================================================================

function getAllPotentialYCorridors(obstacles: Rectangle[]): number[] {
  const yValues = new Set<number>();

  obstacles.forEach((r) => {
    yValues.add(r.top - 25);
    yValues.add(r.bottom + 25);
  });

  if (obstacles.length > 0) {
    let minY = Infinity;
    let maxY = -Infinity;
    obstacles.forEach((r) => {
      minY = Math.min(minY, r.top);
      maxY = Math.max(maxY, r.bottom);
    });
    yValues.add(minY - 60);
    yValues.add(maxY + 60);
  }

  return Array.from(yValues).sort((a, b) => a - b);
}

function getAllPotentialXCorridors(obstacles: Rectangle[]): number[] {
  const xValues = new Set<number>();

  obstacles.forEach((r) => {
    xValues.add(r.left - 25);
    xValues.add(r.right + 25);
  });

  if (obstacles.length > 0) {
    let minX = Infinity;
    let maxX = -Infinity;
    obstacles.forEach((r) => {
      minX = Math.min(minX, r.left);
      maxX = Math.max(maxX, r.right);
    });
    xValues.add(minX - 60);
    xValues.add(maxX + 60);
  }

  return Array.from(xValues).sort((a, b) => a - b);
}

// ============================================================================
// Single Edge Path Building
// ============================================================================

function buildSinglePath(start: Point, end: Point, obstacles: Rectangle[]): Point[] {
  const candidates: Point[][] = [];

  // Straight line
  if (Math.abs(start.y - end.y) < 2 && !hBlocked(start.x, end.x, start.y, obstacles)) {
    return [start, end];
  }
  if (Math.abs(start.x - end.x) < 2 && !vBlocked(start.x, start.y, end.y, obstacles)) {
    return [start, end];
  }

  // L-shaped paths
  const lPathHV = [start, {x: end.x, y: start.y}, end];
  if (isPathClear(lPathHV, obstacles)) {
    candidates.push(lPathHV);
  }

  const lPathVH = [start, {x: start.x, y: end.y}, end];
  if (isPathClear(lPathVH, obstacles)) {
    candidates.push(lPathVH);
  }

  if (candidates.length > 0) {
    candidates.sort((a, b) => pathLength(a) - pathLength(b));
    return candidates[0];
  }

  // 3-segment paths
  const allYCorridors = getAllPotentialYCorridors(obstacles);
  const allXCorridors = getAllPotentialXCorridors(obstacles);

  allYCorridors.forEach((y) => {
    const path = [start, {x: start.x, y}, {x: end.x, y}, end];
    if (isPathClear(path, obstacles)) {
      candidates.push(path);
    }
  });

  allXCorridors.forEach((x) => {
    const path = [start, {x, y: start.y}, {x, y: end.y}, end];
    if (isPathClear(path, obstacles)) {
      candidates.push(path);
    }
  });

  if (candidates.length > 0) {
    candidates.sort((a, b) => pathLength(a) - pathLength(b));
    return candidates[0];
  }

  // 5-segment paths
  allYCorridors.forEach((y) => {
    allXCorridors.forEach((x) => {
      const paths = [
        [start, {x, y: start.y}, {x, y}, {x: end.x, y}, end],
        [start, {x: start.x, y}, {x, y}, {x, y: end.y}, end],
      ];
      paths.forEach((path) => {
        if (isPathClear(path, obstacles)) {
          candidates.push(path);
        }
      });
    });
  });

  if (candidates.length > 0) {
    candidates.sort((a, b) => pathLength(a) - pathLength(b));
    return candidates[0];
  }

  // Fallback
  let minX = Math.min(start.x, end.x);
  let maxX = Math.max(start.x, end.x);
  let minY = Math.min(start.y, end.y);
  let maxY = Math.max(start.y, end.y);

  obstacles.forEach((r) => {
    minX = Math.min(minX, r.left);
    maxX = Math.max(maxX, r.right);
    minY = Math.min(minY, r.top);
    maxY = Math.max(maxY, r.bottom);
  });

  const pad = 80;
  const fallbacks = [
    [start, {x: start.x, y: maxY + pad}, {x: end.x, y: maxY + pad}, end],
    [start, {x: start.x, y: minY - pad}, {x: end.x, y: minY - pad}, end],
    [start, {x: maxX + pad, y: start.y}, {x: maxX + pad, y: end.y}, end],
    [start, {x: minX - pad, y: start.y}, {x: minX - pad, y: end.y}, end],
  ];

  const validFallback = fallbacks.find((path) => isPathClear(path, obstacles));
  if (validFallback) {
    return validFallback;
  }

  return [start, {x: end.x, y: start.y}, end];
}

// ============================================================================
// Segment Extraction and Overlap Detection
// ============================================================================

function extractSegments(edgeId: string, path: Point[]): Segment[] {
  const segments: Segment[] = [];

  for (let i = 0; i < path.length - 1; i += 1) {
    const p1 = path[i];
    const p2 = path[i + 1];

    const isHorizontal = Math.abs(p1.y - p2.y) < 1;
    const isVertical = Math.abs(p1.x - p2.x) < 1;

    if (isHorizontal) {
      segments.push({
        edgeId,
        segmentIndex: i,
        type: 'horizontal',
        p1: {...p1},
        p2: {...p2},
        fixedCoord: p1.y,
        minVar: Math.min(p1.x, p2.x),
        maxVar: Math.max(p1.x, p2.x),
      });
    } else if (isVertical) {
      segments.push({
        edgeId,
        segmentIndex: i,
        type: 'vertical',
        p1: {...p1},
        p2: {...p2},
        fixedCoord: p1.x,
        minVar: Math.min(p1.y, p2.y),
        maxVar: Math.max(p1.y, p2.y),
      });
    }
  }

  return segments;
}

function segmentsOverlap(s1: Segment, s2: Segment): boolean {
  if (s1.type !== s2.type) return false;
  if (s1.edgeId === s2.edgeId) return false;

  // Check if fixed coordinates are close enough
  if (Math.abs(s1.fixedCoord - s2.fixedCoord) > OVERLAP_THRESHOLD) return false;

  // Check if variable ranges overlap
  const overlapStart = Math.max(s1.minVar, s2.minVar);
  const overlapEnd = Math.min(s1.maxVar, s2.maxVar);

  // They overlap if there's a non-trivial overlap region
  return overlapEnd - overlapStart > 5;
}

function findOverlapGroups(allSegments: Segment[]): Segment[][] {
  const groups: Segment[][] = [];
  const processed = new Set<string>();

  const getKey = (s: Segment) => `${s.edgeId}-${s.segmentIndex}`;

  allSegments.forEach((segment) => {
    const key = getKey(segment);
    if (processed.has(key)) return;

    // Find all segments that overlap with this one
    const group: Segment[] = [segment];
    processed.add(key);

    allSegments
      .filter((other) => !processed.has(getKey(other)))
      .forEach((other) => {
        // Check if this segment overlaps with any segment in the group
        const overlapsWithGroup = group.some((g) => segmentsOverlap(g, other));
        if (overlapsWithGroup) {
          group.push(other);
          processed.add(getKey(other));
        }
      });

    if (group.length > 1) {
      groups.push(group);
    }
  });

  return groups;
}

// ============================================================================
// Path Separation
// ============================================================================

interface SegmentOffset {
  edgeId: string;
  segmentIndex: number;
  offset: number; // Perpendicular offset
}

function calculateOffsets(groups: Segment[][]): Map<string, SegmentOffset[]> {
  const offsetMap = new Map<string, SegmentOffset[]>();

  groups.forEach((group) => {
    // Sort group by edge ID for consistent ordering
    group.sort((a, b) => a.edgeId.localeCompare(b.edgeId));

    const count = group.length;
    const totalWidth = (count - 1) * SEPARATION_DISTANCE;
    const startOffset = -totalWidth / 2;

    group.forEach((segment, index) => {
      const offset = startOffset + index * SEPARATION_DISTANCE;

      if (!offsetMap.has(segment.edgeId)) {
        offsetMap.set(segment.edgeId, []);
      }
      offsetMap.get(segment.edgeId)!.push({
        edgeId: segment.edgeId,
        segmentIndex: segment.segmentIndex,
        offset,
      });
    });
  });

  return offsetMap;
}

function applyOffsetsToPath(path: Point[], offsets: SegmentOffset[]): Point[] {
  if (offsets.length === 0) return path;

  // Create a mutable copy
  const newPath = path.map((p) => ({...p}));

  // Sort offsets by segment index
  const sortedOffsets = [...offsets].sort((a, b) => a.segmentIndex - b.segmentIndex);

  sortedOffsets.forEach(({segmentIndex, offset}) => {
    if (segmentIndex >= newPath.length - 1) return;

    const p1 = newPath[segmentIndex];
    const p2 = newPath[segmentIndex + 1];

    const isHorizontal = Math.abs(p1.y - p2.y) < 1;

    if (isHorizontal) {
      // Horizontal segment - offset in Y direction
      p1.y += offset;
      p2.y += offset;
    } else {
      // Vertical segment - offset in X direction
      p1.x += offset;
      p2.x += offset;
    }
  });

  // Reconnect path - ensure consecutive segments share endpoints
  for (let i = 1; i < newPath.length - 1; i += 1) {
    const prev = newPath[i - 1];
    const curr = newPath[i];
    const next = newPath[i + 1];

    // Determine segment types
    const prevToCurrHorizontal = Math.abs(prev.y - curr.y) < 1;
    const currToNextHorizontal = Math.abs(curr.y - next.y) < 1;

    if (prevToCurrHorizontal && !currToNextHorizontal) {
      // Horizontal then vertical - curr.x should match prev-to-curr, curr.y should match curr-to-next
      curr.x = prev.x + (curr.x - prev.x); // Keep x from horizontal segment
    } else if (!prevToCurrHorizontal && currToNextHorizontal) {
      // Vertical then horizontal - curr.y should match prev-to-curr, curr.x should match curr-to-next
      curr.y = prev.y + (curr.y - prev.y); // Keep y from vertical segment
    }
  }

  return newPath;
}

// ============================================================================
// Path Simplification
// ============================================================================

function simplifyPath(points: Point[]): Point[] {
  if (points.length <= 2) return points;

  const result: Point[] = [points[0]];

  for (let i = 1; i < points.length - 1; i += 1) {
    const prev = result[result.length - 1];
    const curr = points[i];
    const next = points[i + 1];

    const collinearH = Math.abs(prev.y - curr.y) < 1 && Math.abs(curr.y - next.y) < 1;
    const collinearV = Math.abs(prev.x - curr.x) < 1 && Math.abs(curr.x - next.x) < 1;

    if (!collinearH && !collinearV) {
      result.push(curr);
    }
  }

  result.push(points[points.length - 1]);
  return result;
}

function removeDuplicates(points: Point[]): Point[] {
  return points.filter((p, i) => {
    if (i === 0) return true;
    return Math.abs(p.x - points[i - 1].x) > 0.5 || Math.abs(p.y - points[i - 1].y) > 0.5;
  });
}

// ============================================================================
// Path to Result Conversion
// ============================================================================

function pathToResult(fullPath: Point[]): EdgePathResult {
  const pathString = fullPath.map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x},${p.y}`).join(' ');

  let totalLength = 0;
  for (let i = 1; i < fullPath.length; i += 1) {
    totalLength += Math.abs(fullPath[i].x - fullPath[i - 1].x) + Math.abs(fullPath[i].y - fullPath[i - 1].y);
  }

  let accumulated = 0;
  let centerX = fullPath[0].x;
  let centerY = fullPath[0].y;

  for (let i = 1; i < fullPath.length; i += 1) {
    const segLen = Math.abs(fullPath[i].x - fullPath[i - 1].x) + Math.abs(fullPath[i].y - fullPath[i - 1].y);

    if (accumulated + segLen >= totalLength / 2) {
      const ratio = segLen > 0 ? (totalLength / 2 - accumulated) / segLen : 0;
      centerX = fullPath[i - 1].x + (fullPath[i].x - fullPath[i - 1].x) * ratio;
      centerY = fullPath[i - 1].y + (fullPath[i].y - fullPath[i - 1].y) * ratio;
      break;
    }
    accumulated += segLen;
  }

  return {path: pathString, centerX, centerY};
}

// ============================================================================
// Main Export - Calculate All Edge Paths with Separation
// ============================================================================

/**
 * Calculate paths for ALL edges with automatic separation of overlapping segments.
 *
 * Usage:
 * 1. Collect all edge data into EdgeInput array
 * 2. Call this function once when edges/nodes change
 * 3. Use the returned Map to get path for each edge by ID
 *
 * @param edges - Array of edge data with positions and handles
 * @param nodes - All nodes in the graph
 * @returns Map of edge ID to EdgePathResult
 */
export function calculateAllEdgePaths(edges: EdgeInput[], nodes: Node[]): Map<string, EdgePathResult> {
  const obstacles = nodes.map((node) => getNodeBounds(node, NODE_MARGIN));
  const results = new Map<string, EdgePathResult>();

  // Step 1: Calculate initial paths for all edges
  const initialPaths = new Map<string, Point[]>();

  edges.forEach((edge) => {
    const sourceExit = getExitPoint(edge.sourceX, edge.sourceY, edge.sourcePosition, obstacles);
    const targetExit = getExitPoint(edge.targetX, edge.targetY, edge.targetPosition, obstacles);

    const middlePath = buildSinglePath(sourceExit, targetExit, obstacles);

    // Assemble full path
    let fullPath: Point[] = [{x: edge.sourceX, y: edge.sourceY}];

    if (Math.abs(sourceExit.x - edge.sourceX) > 0.5 || Math.abs(sourceExit.y - edge.sourceY) > 0.5) {
      fullPath.push(sourceExit);
    }

    middlePath.forEach((p) => {
      const last = fullPath[fullPath.length - 1];
      if (Math.abs(p.x - last.x) > 0.5 || Math.abs(p.y - last.y) > 0.5) {
        fullPath.push(p);
      }
    });

    const lastPoint = fullPath[fullPath.length - 1];
    if (Math.abs(targetExit.x - lastPoint.x) > 0.5 || Math.abs(targetExit.y - lastPoint.y) > 0.5) {
      fullPath.push(targetExit);
    }

    fullPath.push({x: edge.targetX, y: edge.targetY});
    fullPath = simplifyPath(fullPath);
    fullPath = removeDuplicates(fullPath);

    initialPaths.set(edge.id, fullPath);
  });

  // Step 2: Extract all segments
  const allSegments: Segment[] = [];
  initialPaths.forEach((path, edgeId) => {
    allSegments.push(...extractSegments(edgeId, path));
  });

  // Step 3: Find overlapping groups
  const overlapGroups = findOverlapGroups(allSegments);

  // Step 4: Calculate offsets for each overlapping segment
  const offsetMap = calculateOffsets(overlapGroups);

  // Step 5: Apply offsets and generate results
  initialPaths.forEach((path, edgeId) => {
    const edgeOffsets = offsetMap.get(edgeId) ?? [];
    const separatedPath = applyOffsetsToPath(path, edgeOffsets);
    const cleanPath = simplifyPath(separatedPath);
    results.set(edgeId, pathToResult(cleanPath));
  });

  return results;
}

// ============================================================================
// Single Edge Export (for backward compatibility or single edge calculation)
// ============================================================================

/**
 * Calculate path for a single edge (without separation from other edges).
 * Use calculateAllEdgePaths for proper edge separation.
 */
export function calculateEdgePath(
  sourceX: number,
  sourceY: number,
  targetX: number,
  targetY: number,
  sourcePosition: Position,
  targetPosition: Position,
  nodes: Node[],
): EdgePathResult {
  const obstacles = nodes.map((node) => getNodeBounds(node, NODE_MARGIN));

  const sourceExit = getExitPoint(sourceX, sourceY, sourcePosition, obstacles);
  const targetExit = getExitPoint(targetX, targetY, targetPosition, obstacles);

  const middlePath = buildSinglePath(sourceExit, targetExit, obstacles);

  let fullPath: Point[] = [{x: sourceX, y: sourceY}];

  if (Math.abs(sourceExit.x - sourceX) > 0.5 || Math.abs(sourceExit.y - sourceY) > 0.5) {
    fullPath.push(sourceExit);
  }

  middlePath.forEach((p) => {
    const last = fullPath[fullPath.length - 1];
    if (Math.abs(p.x - last.x) > 0.5 || Math.abs(p.y - last.y) > 0.5) {
      fullPath.push(p);
    }
  });

  const lastPoint = fullPath[fullPath.length - 1];
  if (Math.abs(targetExit.x - lastPoint.x) > 0.5 || Math.abs(targetExit.y - lastPoint.y) > 0.5) {
    fullPath.push(targetExit);
  }

  fullPath.push({x: targetX, y: targetY});
  fullPath = simplifyPath(fullPath);
  fullPath = removeDuplicates(fullPath);

  return pathToResult(fullPath);
}
