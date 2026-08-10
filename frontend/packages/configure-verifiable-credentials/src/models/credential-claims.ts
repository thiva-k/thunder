// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import type {ClaimMapping, VerifiableCredential} from './vc';

let rowCounter = 0;

/** An editable claim row: the attribute name and its display name. */
export interface ClaimRow {
  id: string;
  name: string;
  displayName: string;
}

/** emptyClaimRow returns a fresh, blank claim row with a stable local id. */
export function emptyClaimRow(): ClaimRow {
  rowCounter += 1;
  return {id: `claim-${rowCounter}`, name: '', displayName: ''};
}

/** credentialToClaimRows builds editor rows from a credential's claims. */
export function credentialToClaimRows(credential?: VerifiableCredential): ClaimRow[] {
  const claims = credential?.claims ?? [];
  if (claims.length === 0) {
    return [emptyClaimRow()];
  }
  return claims.map((c: ClaimMapping): ClaimRow => {
    rowCounter += 1;
    return {id: `claim-${rowCounter}`, name: c.name, displayName: c.displayName ?? ''};
  });
}

/** claimRowsToRequest maps editor rows to the API claims array, dropping unnamed rows. */
export function claimRowsToRequest(rows: ClaimRow[]): ClaimMapping[] {
  return rows
    .filter((r: ClaimRow): boolean => r.name.trim() !== '')
    .map(
      (r: ClaimRow): ClaimMapping => ({
        name: r.name.trim(),
        displayName: r.displayName.trim() || undefined,
      }),
    );
}
