/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

import React from 'react';

// Reusable person-silhouette icon. PersonIcon viewBox is 56x56; the
// returned <g> renders within whatever transform / scale the caller
// applies. The outer circle is themed via the `className` prop; the
// inner glyph uses a fixed class so the same person silhouette renders
// in both diagrams.
function PersonIcon({className}: {className?: string}) {
  return (
    <g className={className}>
      <circle cx="28" cy="28" r="26" />
      <g transform="translate(28,28)" className="uc-b2c-wayfinder-person-glyph">
        <circle cx="0" cy="-6" r="7" />
        <path d="M -13 14 C -13 4 13 4 13 14 Z" />
      </g>
    </g>
  );
}

/**
 * "Meet Wayfinder" diagram. The top card names the app; the two
 * columns below show consumers (peers, stacked) and staff (Odin at
 * the top with hierarchy connectors down to Heimdall and Baldur).
 * Hierarchy lines align with the staff icons' horizontal centres.
 */
export function WayfinderOrganization() {
  return (
    <div className="uc-b2c-wayfinder-org">
      <svg
        className="uc-b2c-wayfinder-org__svg"
        viewBox="0 0 960 660"
        xmlns="http://www.w3.org/2000/svg"
        role="img"
        aria-label="Wayfinder organization structure"
      >
        {/* Wayfinder header */}
        <g className="uc-b2c-wayfinder-org__header">
          <rect x="280" y="20" width="400" height="80" rx="12" />
          <text x="480" y="52" textAnchor="middle" className="uc-b2c-wayfinder-org__header-title">
            Wayfinder
          </text>
          <text x="480" y="78" textAnchor="middle" className="uc-b2c-wayfinder-org__header-sub">
            B2C Travel-Booking Application
          </text>
        </g>

        {/* Trunk connectors splitting into two columns */}
        <g className="uc-b2c-wayfinder-org__edges">
          <line x1="480" y1="100" x2="480" y2="122" />
          <line x1="260" y1="122" x2="700" y2="122" />
          <line x1="260" y1="122" x2="260" y2="148" />
          <line x1="700" y1="122" x2="700" y2="148" />
        </g>

        {/* Consumers column — peers, stacked vertically */}
        <g className="uc-b2c-wayfinder-org__col uc-b2c-wayfinder-org__col--consumers" transform="translate(80,148)">
          <rect width="360" height="492" rx="10" />
          <text x="180" y="38" textAnchor="middle" className="uc-b2c-wayfinder-org__col-title">
            Consumers
          </text>
          <text x="180" y="60" textAnchor="middle" className="uc-b2c-wayfinder-org__col-sub">
            Book travel
          </text>
          <line x1="40" y1="78" x2="320" y2="78" className="uc-b2c-wayfinder-org__divider" />

          {/* Thor (top) */}
          <g transform="translate(156,95)">
            <g transform="scale(0.86)">
              <PersonIcon className="uc-b2c-wayfinder-org__icon" />
            </g>
          </g>
          <text x="180" y="171" textAnchor="middle" className="uc-b2c-wayfinder-org__cast-name">
            Thor
          </text>
          <text x="180" y="191" textAnchor="middle" className="uc-b2c-wayfinder-org__cast-role">
            Returning traveller
          </text>

          {/* Sif (middle) */}
          <g transform="translate(156,215)">
            <g transform="scale(0.86)">
              <PersonIcon className="uc-b2c-wayfinder-org__icon" />
            </g>
          </g>
          <text x="180" y="291" textAnchor="middle" className="uc-b2c-wayfinder-org__cast-name">
            Sif
          </text>
          <text x="180" y="311" textAnchor="middle" className="uc-b2c-wayfinder-org__cast-role">
            New arrival via email
          </text>

          {/* Freya (bottom) */}
          <g transform="translate(156,335)">
            <g transform="scale(0.86)">
              <PersonIcon className="uc-b2c-wayfinder-org__icon" />
            </g>
          </g>
          <text x="180" y="411" textAnchor="middle" className="uc-b2c-wayfinder-org__cast-name">
            Freya
          </text>
          <text x="180" y="431" textAnchor="middle" className="uc-b2c-wayfinder-org__cast-role">
            New arrival via Google
          </text>
        </g>

        {/* Staff column — hierarchy */}
        <g className="uc-b2c-wayfinder-org__col uc-b2c-wayfinder-org__col--staff" transform="translate(520,148)">
          <rect width="360" height="492" rx="10" />
          <text x="180" y="38" textAnchor="middle" className="uc-b2c-wayfinder-org__col-title">
            Staff
          </text>
          <text x="180" y="60" textAnchor="middle" className="uc-b2c-wayfinder-org__col-sub">
            Run the product
          </text>
          <line x1="40" y1="78" x2="320" y2="78" className="uc-b2c-wayfinder-org__divider" />

          {/* Odin (centered at the top of the column) */}
          <g transform="translate(156,110)">
            <g transform="scale(0.86)">
              <PersonIcon className="uc-b2c-wayfinder-org__icon uc-b2c-wayfinder-org__icon--lead" />
            </g>
          </g>
          <text x="180" y="186" textAnchor="middle" className="uc-b2c-wayfinder-org__cast-name">
            Odin
          </text>
          <text x="180" y="206" textAnchor="middle" className="uc-b2c-wayfinder-org__cast-role">
            Operations admin
          </text>

          {/* Hierarchy connector — all lines aligned with icon centres */}
          <g className="uc-b2c-wayfinder-org__edges">
            <line x1="180" y1="220" x2="180" y2="252" />
            <line x1="110" y1="252" x2="250" y2="252" />
            <line x1="110" y1="252" x2="110" y2="280" />
            <line x1="250" y1="252" x2="250" y2="280" />
          </g>

          {/* Heimdall (left report, icon centre at x=110) */}
          <g transform="translate(86,280)">
            <g transform="scale(0.86)">
              <PersonIcon className="uc-b2c-wayfinder-org__icon" />
            </g>
          </g>
          <text x="110" y="356" textAnchor="middle" className="uc-b2c-wayfinder-org__cast-name">
            Heimdall
          </text>
          <text x="110" y="376" textAnchor="middle" className="uc-b2c-wayfinder-org__cast-role">
            Support agent
          </text>

          {/* Baldur (right report, icon centre at x=250) */}
          <g transform="translate(226,280)">
            <g transform="scale(0.86)">
              <PersonIcon className="uc-b2c-wayfinder-org__icon" />
            </g>
          </g>
          <text x="250" y="356" textAnchor="middle" className="uc-b2c-wayfinder-org__cast-name">
            Baldur
          </text>
          <text x="250" y="376" textAnchor="middle" className="uc-b2c-wayfinder-org__cast-role">
            Destinations curator
          </text>
        </g>
      </svg>
    </div>
  );
}

/**
 * Architecture diagram. Consumers (Thor, Freya) sit at the top next
 * to the Wayfinder Web app; ThunderID and Wayfinder Server sit below
 * the app, symmetrically. Pattern-agnostic — the arrow labels do not
 * commit to redirect-based vs app-native vs direct API.
 */
export function WayfinderArchitecture() {
  return (
    <div className="uc-b2c-wayfinder-arch">
      <svg
        className="uc-b2c-wayfinder-arch__svg"
        viewBox="0 0 960 720"
        xmlns="http://www.w3.org/2000/svg"
        role="img"
        aria-label="Wayfinder app, server, and ThunderID integration"
      >
        <defs>
          <marker
            id="uc-b2c-arch-arrow"
            viewBox="0 0 10 10"
            refX="9"
            refY="5"
            markerWidth="6"
            markerHeight="6"
            orient="auto-start-reverse"
          >
            <path d="M0,0 L10,5 L0,10 z" fill="currentColor" />
          </marker>
        </defs>

        {/* Consumers — top, near the Wayfinder Web app */}
        <g className="uc-b2c-wayfinder-arch__consumers">
          <text x="480" y="32" textAnchor="middle" className="uc-b2c-wayfinder-arch__group-label">
            Consumers
          </text>

          {/* Thor */}
          <g transform="translate(358,46)">
            <g transform="scale(0.78)">
              <PersonIcon className="uc-b2c-wayfinder-arch__icon" />
            </g>
          </g>
          <text x="380" y="116" textAnchor="middle" className="uc-b2c-wayfinder-arch__cast-name">
            Thor
          </text>

          {/* Sif */}
          <g transform="translate(458,46)">
            <g transform="scale(0.78)">
              <PersonIcon className="uc-b2c-wayfinder-arch__icon" />
            </g>
          </g>
          <text x="480" y="116" textAnchor="middle" className="uc-b2c-wayfinder-arch__cast-name">
            Sif
          </text>

          {/* Freya */}
          <g transform="translate(558,46)">
            <g transform="scale(0.78)">
              <PersonIcon className="uc-b2c-wayfinder-arch__icon" />
            </g>
          </g>
          <text x="580" y="116" textAnchor="middle" className="uc-b2c-wayfinder-arch__cast-name">
            Freya
          </text>
        </g>

        {/* Arrow from consumers down to Wayfinder Web */}
        <g className="uc-b2c-wayfinder-arch__edges">
          <line x1="480" y1="130" x2="480" y2="170" markerEnd="url(#uc-b2c-arch-arrow)" />
          <text x="494" y="156" className="uc-b2c-wayfinder-arch__edge-label">
            use
          </text>
        </g>

        {/* Wayfinder Web — middle */}
        <g className="uc-b2c-wayfinder-arch__app" transform="translate(290,180)">
          <rect width="380" height="130" rx="12" />
          <text x="190" y="40" textAnchor="middle" className="uc-b2c-wayfinder-arch__app-title">
            Wayfinder Web
          </text>
          <text x="190" y="64" textAnchor="middle" className="uc-b2c-wayfinder-arch__sub">
            Browser-based SPA
          </text>
          <line x1="40" y1="80" x2="340" y2="80" className="uc-b2c-wayfinder-arch__divider" />
          <text x="190" y="104" textAnchor="middle" className="uc-b2c-wayfinder-arch__detail">
            Book travel
          </text>
        </g>

        {/* ThunderID — bottom left */}
        <g className="uc-b2c-wayfinder-arch__idp" transform="translate(80,400)">
          <rect width="320" height="140" rx="12" />
          <text x="160" y="40" textAnchor="middle" className="uc-b2c-wayfinder-arch__idp-title">
            ThunderID
          </text>
          <text x="160" y="64" textAnchor="middle" className="uc-b2c-wayfinder-arch__sub">
            Identity Authority
          </text>
          <line x1="40" y1="80" x2="280" y2="80" className="uc-b2c-wayfinder-arch__divider" />
          <text x="160" y="104" textAnchor="middle" className="uc-b2c-wayfinder-arch__detail">
            Manages users, issues tokens
          </text>
        </g>

        {/* Wayfinder Server — bottom right */}
        <g className="uc-b2c-wayfinder-arch__app" transform="translate(560,400)">
          <rect width="320" height="140" rx="12" />
          <text x="160" y="40" textAnchor="middle" className="uc-b2c-wayfinder-arch__app-title">
            Wayfinder Server
          </text>
          <text x="160" y="64" textAnchor="middle" className="uc-b2c-wayfinder-arch__sub">
            Booking API
          </text>
          <line x1="40" y1="80" x2="280" y2="80" className="uc-b2c-wayfinder-arch__divider" />
          <text x="160" y="104" textAnchor="middle" className="uc-b2c-wayfinder-arch__detail">
            Holds bookings, flights, hotels
          </text>
        </g>

        {/* Arrows from Wayfinder Web to ThunderID and Server */}
        <g className="uc-b2c-wayfinder-arch__edges">
          {/* Wayfinder Web ↔ ThunderID */}
          <line x1="380" y1="310" x2="240" y2="400" markerEnd="url(#uc-b2c-arch-arrow)" />
          <line x1="220" y1="400" x2="360" y2="310" markerEnd="url(#uc-b2c-arch-arrow)" />
          <text x="232" y="346" className="uc-b2c-wayfinder-arch__edge-label">
            Sign-in,
          </text>
          <text x="232" y="362" className="uc-b2c-wayfinder-arch__edge-label">
            sign-up, recovery
          </text>

          {/* Wayfinder Web ↔ Wayfinder Server */}
          <line x1="580" y1="310" x2="720" y2="400" markerEnd="url(#uc-b2c-arch-arrow)" />
          <line x1="740" y1="400" x2="600" y2="310" markerEnd="url(#uc-b2c-arch-arrow)" />
          <text x="666" y="346" className="uc-b2c-wayfinder-arch__edge-label">
            Authenticated
          </text>
          <text x="666" y="362" className="uc-b2c-wayfinder-arch__edge-label">
            API calls
          </text>
        </g>

        {/* Arrow from staff up to ThunderID */}
        <g className="uc-b2c-wayfinder-arch__edges">
          <line x1="240" y1="590" x2="240" y2="550" markerEnd="url(#uc-b2c-arch-arrow)" />
          <text x="254" y="576" className="uc-b2c-wayfinder-arch__edge-label">
            Console
          </text>
        </g>

        {/* Staff — below ThunderID, mirroring consumers above Wayfinder Web */}
        <g className="uc-b2c-wayfinder-arch__consumers">
          <text x="240" y="610" textAnchor="middle" className="uc-b2c-wayfinder-arch__group-label">
            Staff
          </text>

          {/* Odin */}
          <g transform="translate(118,624)">
            <g transform="scale(0.78)">
              <PersonIcon className="uc-b2c-wayfinder-arch__icon" />
            </g>
          </g>
          <text x="140" y="694" textAnchor="middle" className="uc-b2c-wayfinder-arch__cast-name">
            Odin
          </text>

          {/* Heimdall */}
          <g transform="translate(218,624)">
            <g transform="scale(0.78)">
              <PersonIcon className="uc-b2c-wayfinder-arch__icon" />
            </g>
          </g>
          <text x="240" y="694" textAnchor="middle" className="uc-b2c-wayfinder-arch__cast-name">
            Heimdall
          </text>

          {/* Baldur */}
          <g transform="translate(318,624)">
            <g transform="scale(0.78)">
              <PersonIcon className="uc-b2c-wayfinder-arch__icon" />
            </g>
          </g>
          <text x="340" y="694" textAnchor="middle" className="uc-b2c-wayfinder-arch__cast-name">
            Baldur
          </text>
        </g>
      </svg>
    </div>
  );
}
