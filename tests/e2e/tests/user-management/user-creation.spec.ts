// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

/**
 * User Management E2E Tests
 *
 * Tests for user CRUD operations in the Console.
 * Uses Page Object Model pattern via fixtures.
 *
 * Required environment variables:
 * - BASE_URL: Console base URL
 * - TEST_USER_USERNAME: Base username for test user creation
 * - ADMIN_USERNAME: Admin credentials for authentication
 * - ADMIN_PASSWORD: Admin password for authentication
 */

import { test, type UserFormData } from "../../fixtures/console";

const baseUsername = process.env.TEST_USER_USERNAME as string;

/**
 * Generates unique test data for user creation
 * @param suffix - Optional suffix to identify test case
 */
const generateTestData = (suffix: string = ""): UserFormData => {
  const timestamp = Date.now();
  const randomSuffix = Math.floor(Math.random() * 1000);
  const uniqueSuffix = suffix ? `${suffix}_${timestamp}_${randomSuffix}` : `${timestamp}_${randomSuffix}`;

  return {
    username: `${baseUsername}${uniqueSuffix}`,
    email: `${baseUsername}${uniqueSuffix}@wso2.com`,
    given_name: `Testfname${suffix}`,
    family_name: `Testlname${suffix}`,
  };
};

test.describe("User Management - CRUD Operations", () => {
  test.describe("Create User Operations", () => {
    /** TC001: Verify user can be created with all required fields */
    test("TC001: Create new user with all required fields", async ({ usersPage }) => {
      const testData = generateTestData("001");

      await test.step("Navigate to Create User Wizard", async () => {
        console.log("Navigating directly to create user wizard...");
        await usersPage.gotoCreateUserWizard();
        console.log("Successfully accessed create user wizard");
        await usersPage.screenshot("debug-create-user-wizard-loaded");
      });

      await test.step("Select user type and continue", async () => {
        console.log("Selecting user type...");
        await usersPage.selectUserTypeAndContinue();
        console.log("User type selected, advanced to details step");
        await usersPage.screenshot("debug-user-details-step");
      });

      await test.step("Fill in user details", async () => {
        console.log("Filling user details:", testData);
        await usersPage.fillUserForm(testData);
        console.log("User details filled");
        await usersPage.screenshot("debug-form-filled");
      });

      await test.step("Submit user creation form", async () => {
        console.log("Submitting user creation form...");
        await usersPage.submitForm();
        console.log("User creation form submitted");
        await usersPage.page.waitForLoadState("networkidle");
        await usersPage.screenshot("debug-after-creation");
      });
    });
  });
});
