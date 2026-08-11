// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

/**
 * Users Page Object Model
 *
 * Encapsulates all locators and actions for the User Management page.
 *
 * @example
 * const usersPage = new UsersPage(page, baseUrl);
 * await usersPage.goto();
 * await usersPage.createUser({ username: 'test', email: 'test@test.com' });
 */

import { Page, Locator, expect } from "@playwright/test";
import { ConsoleRoutes } from "../../configs/routes/console-routes";
import { BasePage } from "../base.page";
import { Timeouts } from "../../constants/timeouts";

export type UserFormData = {
  username: string;
  email: string;
  given_name?: string;
  family_name?: string;
};

export class UsersPage extends BasePage {
  readonly baseUrl: string;

  // Page Locators
  readonly addUserButton: Locator;
  readonly userTable: Locator;
  readonly searchInput: Locator;

  // Wizard Locators (Step 1: Select User Type)
  readonly userTypeHeading: Locator;
  readonly organizationUnitHeading: Locator;
  readonly onboardingModeHeading: Locator;
  readonly userTypeSelect: Locator;
  readonly continueButton: Locator;
  readonly createUserActionButton: Locator;

  // Form Locators (Step 2: User Details)
  readonly usernameInput: Locator;
  readonly emailInput: Locator;
  readonly givenNameInput: Locator;
  readonly familyNameInput: Locator;
  readonly passwordInput: Locator;
  readonly submitButton: Locator;
  readonly cancelButton: Locator;
  readonly formHeading: Locator;

  // Messages
  readonly successMessage: Locator;
  readonly errorMessage: Locator;

  constructor(page: Page, baseUrl: string) {
    super(page);
    this.baseUrl = baseUrl;

    // Add User button
    this.addUserButton = page
      .getByRole("button", { name: /add user/i })
      .or(page.locator('button:has-text("Add User")'))
      .or(page.locator('button:has-text("+ Add User")'))
      .or(page.locator('[data-testid*="add"][data-testid*="user"]'))
      .or(page.locator('a:has-text("Add User")'));

    // Wizard: Step 1 heading ("Select a user type")
    this.userTypeHeading = page.locator("h1, h2, h3, h4, h5, h6").filter({ hasText: /select.*user.*type/i });
    this.organizationUnitHeading = page
      .locator("h1, h2, h3, h4, h5, h6")
      .filter({ hasText: /select an organization unit/i });
    this.onboardingModeHeading = page.locator("h1, h2, h3, h4, h5, h6").filter({ hasText: /^add user$/i });

    // Wizard: User type dropdown
    this.userTypeSelect = page
      .locator('[data-testid="user-type-select"]')
      .or(page.locator("#user-type-select"))
      .or(page.getByRole("combobox"))
      .or(page.locator('[aria-haspopup="listbox"]'));

    // Wizard: Continue button
    this.continueButton = page.getByRole("button", { name: /continue/i });
    this.createUserActionButton = page.getByRole("button", { name: /^create user$/i });

    // User table
    this.userTable = page.locator('table, [role="table"], [data-testid*="user-list"]');

    // Search input
    this.searchInput = page.locator('input[placeholder*="search" i], input[type="search"]');

    // Form fields - support both embedded flow (by id/label) and traditional form (by name)
    this.usernameInput = page.locator('input#username')
      .or(page.locator('input[name="username"]'))
      .or(page.getByLabel(/username/i));

    this.emailInput = page.locator('input#email')
      .or(page.locator('input[name="email"]'))
      .or(page.getByLabel(/email/i));

    this.givenNameInput = page.locator('input#given_name')
      .or(page.locator('input[name="given_name"]'))
      .or(page.getByLabel(/first.*name|given.*name/i));

    this.familyNameInput = page.locator('input#family_name')
      .or(page.locator('input[name="family_name"]'))
      .or(page.getByLabel(/last.*name|family.*name/i));
    this.passwordInput = page.locator('input#password')
      .or(page.locator('input[name="password"]'))
      .or(page.getByLabel(/^password$/i));

    // Form buttons - support multiple button naming conventions
    // The actual button label comes from the embedded flow, so we support common patterns
    this.submitButton = page
      .getByRole("button", { name: /create.*user|add.*user|submit|save|finish|next|continue|confirm/i })
      .or(page.locator('button[size="large"]:not(:has-text("Cancel"))')
        .or(page.locator('button:not(:has-text("Cancel")):not(:has-text("Close"))').nth(-1)));
    this.cancelButton = page.getByRole("button", { name: /cancel|close/i });

    // Form heading (Step 2: "Enter user details")
    this.formHeading = page
      .locator("h1, h2, h3, h4, h5, h6")
      .filter({ hasText: /enter.*user.*details|user.*details/i });

    // Messages
    this.successMessage = page.locator('[class*="success"], [role="status"]');
    this.errorMessage = page.locator('[class*="error"], [role="alert"]');
  }

  /** Navigate to users management page */
  async goto() {
    await this.page.goto(`${this.baseUrl}${ConsoleRoutes.users}`, {
      waitUntil: "networkidle",
      timeout: Timeouts.PAGE_LOAD,
    });
  }

  /** Navigate directly to create user wizard (bypassing selection page) */
  async gotoCreateUserWizard() {
    await this.page.goto(`${this.baseUrl}${ConsoleRoutes.users}/add/create`, {
      waitUntil: "networkidle",
      timeout: Timeouts.PAGE_LOAD,
    });
    // Wait for the wizard to fully load
    await this.waitForUserForm();
  }

  /** Check if currently on users page */
  async isOnUsersPage(): Promise<boolean> {
    const url = this.page.url();
    return url.includes(ConsoleRoutes.users) && !url.includes(ConsoleRoutes.signin);
  }

  /** Verify page loaded successfully */
  async verifyPageLoaded() {
    const url = this.page.url();
    if (url.includes(ConsoleRoutes.signin)) {
      throw new Error("Authentication failed: Redirected to signin page");
    }
    expect(url).toContain(ConsoleRoutes.users);
  }

  /** Click the Add User button */
  async clickAddUser() {
    await this.addUserButton.first().waitFor({ state: "visible", timeout: Timeouts.ELEMENT_VISIBILITY });
    await this.addUserButton.first().scrollIntoViewIfNeeded();
    await this.addUserButton.first().click();
  }

  /** Click the Create User option card on the Add User selection page */
  async clickCreateUserOption() {
    const createUserCard = this.page
      .locator('[data-testid="add-user-type-select"]')
      .locator(':has-text("Create User")');

    await createUserCard.first().waitFor({ state: "visible", timeout: Timeouts.ELEMENT_VISIBILITY });
    await createUserCard.first().click();
  }

  /** Wait for the wizard to load (Step 1: Select User Type) */
  async waitForUserForm() {
    await this.waitForAnyVisibleLocator(
      [this.userTypeHeading, this.organizationUnitHeading, this.onboardingModeHeading, this.formHeading],
      Timeouts.FORM_LOAD,
    );
  }

  /** Select the first available user type and advance to Step 2 */
  async selectUserTypeAndContinue() {
    // Wait for user type select to be visible
    await this.userTypeSelect.first().waitFor({ state: "visible", timeout: Timeouts.FORM_LOAD });

    // Click the user type select dropdown
    await this.userTypeSelect.first().click();

    // Select the first available option
    const firstOption = this.page.locator('[role="option"]:not([aria-disabled="true"])').first();
    await firstOption.waitFor({ state: "visible", timeout: Timeouts.ELEMENT_VISIBILITY });
    await firstOption.click();

    // Click Continue button
    await this.continueButton.first().waitFor({ state: "visible", timeout: Timeouts.ELEMENT_VISIBILITY });
    await this.clickContinueButton();

    // Wait for the page to transition - wait until the user type heading disappears or changes
    try {
      await this.userTypeHeading.first().waitFor({ state: "hidden", timeout: Timeouts.FORM_LOAD });
    } catch {
      // Heading might not disappear, just wait for page to settle
      await this.page.waitForLoadState("networkidle", {timeout: Timeouts.FORM_LOAD}).catch(() => {});
    }

    // Wait a moment for animations to complete
    await this.page.waitForTimeout(300);

    // Handle Organization Unit step if it appears
    if (await this.isLocatorVisible(this.organizationUnitHeading)) {
      await this.continueButton.first().waitFor({ state: "visible", timeout: Timeouts.ELEMENT_VISIBILITY });
      await this.clickContinueButton();

      // Wait for organization unit heading to disappear
      try {
        await this.organizationUnitHeading.first().waitFor({ state: "hidden", timeout: Timeouts.FORM_LOAD });
      } catch {
        await this.page.waitForLoadState("networkidle", {timeout: Timeouts.FORM_LOAD}).catch(() => {});
      }
      await this.page.waitForTimeout(300);
    }

    // Handle Create User action button if it appears (onboarding mode selection)
    if (await this.isLocatorVisible(this.createUserActionButton)) {
      await this.createUserActionButton.first().click();

      // Wait for page to settle after clicking
      try {
        await this.createUserActionButton.first().waitFor({ state: "hidden", timeout: Timeouts.FORM_LOAD });
      } catch {
        await this.page.waitForLoadState("networkidle", {timeout: Timeouts.FORM_LOAD}).catch(() => {});
      }
      await this.page.waitForTimeout(300);
    }

    // Wait for details step to load
    await this.waitForDetailsStep();
  }

  /** Fill the user form (Step 2: User Details) */
  async fillUserForm(data: UserFormData) {
    // Fill known fields by name/label
    if (
      await this.usernameInput
        .first()
        .isVisible()
        .catch(() => false)
    ) {
      await this.usernameInput.first().fill(data.username);
    }
    if (
      await this.emailInput
        .first()
        .isVisible()
        .catch(() => false)
    ) {
      await this.emailInput.first().fill(data.email);
    }
    if (
      data.given_name &&
      (await this.givenNameInput
        .first()
        .isVisible()
        .catch(() => false))
    ) {
      await this.givenNameInput.first().fill(data.given_name);
    }
    if (
      data.family_name &&
      (await this.familyNameInput
        .first()
        .isVisible()
        .catch(() => false))
    ) {
      await this.familyNameInput.first().fill(data.family_name);
    }

    // Fill any remaining empty required text/password inputs with generated values
    // (dynamic schema fields that aren't covered by the known field locators)
    const requiredInputs = this.page.locator('input[required]:not([type="checkbox"]):not([type="radio"])');
    const count = await requiredInputs.count();
    for (let i = 0; i < count; i++) {
      const input = requiredInputs.nth(i);
      const currentValue = await input.inputValue();
      if (!currentValue) {
        const name = (await input.getAttribute("name")) ?? `field_${i}`;
        const type = await input.getAttribute("type");
        const value = type === "password" ? `Test@${Date.now()}` : `test_${name}_${Date.now()}`;
        await input.fill(value);
      }
    }
  }

  /** Submit the form (clicks "Create User" on the last step) */
  async submitForm() {
    const submitBtn = this.submitButton.first();

    // Wait for button to be visible first
    try {
      await submitBtn.waitFor({ state: "visible", timeout: Timeouts.ELEMENT_VISIBILITY });
    } catch (error) {
      // If button not found by standard selectors, try to find any large contained button
      const allButtons = this.page.locator('button[size="large"]');
      const count = await allButtons.count();

      if (count > 0) {
        // Try the last button (usually submit in wizards)
        await allButtons.last().waitFor({ state: "visible", timeout: Timeouts.ELEMENT_VISIBILITY });
        await allButtons.last().click();
        return;
      }

      throw new Error(`Could not find submit button. ${error}`);
    }

    // Check if enabled and click
    await expect(submitBtn).toBeEnabled({ timeout: Timeouts.ELEMENT_VISIBILITY });
    await submitBtn.click();
  }

  private async clickContinueButton() {
    await this.continueButton.first().waitFor({ state: "visible", timeout: Timeouts.ELEMENT_VISIBILITY });
    await this.continueButton.first().click();
  }

  private async waitForDetailsStep() {
    // Strategy 1: Wait for specific typed text input fields to be visible (form is interactive)
    const typedInputs = this.page.locator('input[type="text"], input[type="email"], input[type="password"], input[type="tel"], textarea');

    try {
      await typedInputs.first().waitFor({ state: "visible", timeout: Timeouts.FORM_LOAD / 2 });
      return;
    } catch {
      // Strategy 2: Wait for ANY input element or form field
      const anyInputs = this.page.locator('input:not([type="hidden"]):not([type="checkbox"]):not([type="radio"])');
      const anyTextfields = this.page.locator('[role="textbox"]');
      const formControl = this.page.locator('[role="group"], .MuiFormControl-root');

      try {
        await Promise.race([
          anyInputs.first().waitFor({ state: "visible", timeout: Timeouts.FORM_LOAD / 2 }),
          anyTextfields.first().waitFor({ state: "visible", timeout: Timeouts.FORM_LOAD / 2 }),
          formControl.first().waitFor({ state: "visible", timeout: Timeouts.FORM_LOAD / 2 }),
        ]);
        return;
      } catch {
        // Strategy 3: Fallback with detailed error reporting
        await this.waitForAnyVisibleLocator(
          [this.formHeading, this.usernameInput, this.emailInput, this.givenNameInput, this.familyNameInput, this.passwordInput, anyInputs, anyTextfields],
          Timeouts.FORM_LOAD,
        );
      }
    }
  }

  private async isLocatorVisible(locator: Locator): Promise<boolean> {
    return locator.first().isVisible();
  }

  private async waitForAnyVisibleLocator(locators: Locator[], timeout: number) {
    try {
      await Promise.any(
        locators.map((locator) => locator.first().waitFor({ state: "visible", timeout })),
      );
    } catch (error) {
      // Provide debug information about what's actually on the page
      const pageContent = await this.page.content();
      const hasInputs = pageContent.includes('<input');
      const hasFormControl = pageContent.includes('FormControl');
      const hasTextField = pageContent.includes('TextField');

      // Try to find any inputs on the page and log their details
      const allInputs = await this.page.locator('input').all();
      const inputDetails = await Promise.all(
        allInputs.slice(0, 5).map(async (input) => {
          try {
            const type = await input.getAttribute('type');
            const id = await input.getAttribute('id');
            const name = await input.getAttribute('name');
            const visible = await input.isVisible().catch(() => false);
            return {type, id, name, visible};
          } catch {
            return null;
          }
        }),
      );

      const headingContent = await this.page.locator('h1, h2, h3, h4, h5, h6').first().textContent().catch(() => '');

      throw new Error(
        `Timed out after ${timeout}ms while waiting for the next visible user-creation step. ` +
        `Debug: hasInputs=${hasInputs}, hasFormControl=${hasFormControl}, hasTextField=${hasTextField}. ` +
        `Found ${allInputs.length} inputs: ${JSON.stringify(inputDetails.filter(Boolean))}. ` +
        `Heading: "${headingContent}". ` +
        `Error: ${error instanceof Error ? error.message : String(error)}`
      );
    }
  }

  /** Cancel the form */
  async cancelForm() {
    await this.cancelButton.first().click();
  }

  /** Create a new user (complete wizard flow) */
  async createUser(data: UserFormData) {
    await this.clickAddUser();
    await this.waitForUserForm();
    await this.selectUserTypeAndContinue();
    await this.fillUserForm(data);
    await this.submitForm();
  }

  /** Search for a user */
  async searchUser(query: string) {
    await this.searchInput.first().fill(query);
    // Using network idle after triggering search.
    // This is acceptable here because the users page is expected not to keep long-lived
    // connections (e.g., websockets) and search is the primary network activity.
    // If additional long-running requests are introduced, prefer a more targeted wait
    // such as page.waitForResponse() for the search API or waiting for the results
    // table locator to update instead of relying on 'networkidle'.
    await this.page.waitForLoadState("networkidle");
  }

  /** Get user count */
  async getUserCount(): Promise<number> {
    const rows = this.page.locator('table tbody tr, [role="row"]');
    return await rows.count();
  }
}
