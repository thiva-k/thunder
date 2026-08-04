// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

/**
 * Sample App Fixture
 *
 * Provides page object instances for sample app testing.
 */

import { test as base } from "@playwright/test";
import { SampleAppLoginPage } from "../../pages/sample-app";

type SampleAppFixtures = {
  sampleAppLoginPage: SampleAppLoginPage;
};

/**
 * Extended test fixture with sample app page objects.
 */
export const test = base.extend<SampleAppFixtures>({
  /**
   * Sample App Login Page fixture
   */
  sampleAppLoginPage: async ({ page }, use) => {
    const sampleAppLoginPage = new SampleAppLoginPage(page);
    await use(sampleAppLoginPage);
  },
});

export { expect } from "@playwright/test";
