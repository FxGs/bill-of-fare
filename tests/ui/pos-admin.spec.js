const { test, expect } = require("@playwright/test");

test.describe("POS browser flows", () => {
  test("adds a menu item, creates an order, and shows the invoice preview", async ({ page }) => {
    await page.goto("/");

    await expect(page.getByRole("heading", { name: "Bill of Fare" })).toBeVisible();
    await expect(page.getByRole("link", { name: "Admin" })).toBeVisible();

    const sandwichCard = page.locator("article.menu-card").filter({ hasText: "Paneer Cheese Grilled Sandwich" });
    await sandwichCard.getByRole("button", { name: /Add Paneer Cheese Grilled Sandwich/ }).click();

    await expect(page.locator(".order-pane")).toContainText("Paneer Cheese Grilled Sandwich");
    await expect(page.locator(".order-pane")).toContainText("₹130");

    await page.getByRole("button", { name: "Create Order" }).click();
    await expect(page.getByRole("heading", { name: /Create order #/ })).toBeVisible();
    await page.getByRole("button", { name: "Yes, Create Order" }).click();

    await expect(page.getByRole("heading", { name: "Invoice Preview" })).toBeVisible();
    await expect(page.locator(".invoice-preview-receipt")).toContainText("Paneer Cheese Grilled Sandwich");
    await expect(page.locator(".invoice-preview-receipt")).toContainText("₹130");
    await expect(page.getByRole("link", { name: "Print" })).toBeVisible();
    await expect(page.locator(".order-pane")).toContainText("No items added yet");
  });

  test("opens the variant chooser before adding a variant item", async ({ page }) => {
    await page.goto("/");

    await page.locator(".category-tabs").getByRole("button", { name: "Roll", exact: true }).click();
    const rollCard = page.locator("article.menu-card").filter({
      has: page.getByRole("heading", { name: "Egg Roll", exact: true }),
    });
    await expect(rollCard).toContainText("2 variants");
    await rollCard.getByRole("button", { name: /Choose Egg Roll variant/ }).click();

    const variantModal = page.getByRole("dialog", { name: "Egg Roll", exact: true });
    await expect(variantModal).toBeVisible();
    await variantModal.getByRole("button", { name: /Add Egg Roll Cheese/ }).click();

    await expect(variantModal).not.toBeVisible();
    await expect(page.locator(".order-pane")).toContainText("Egg Roll");
    await expect(page.locator(".order-pane")).toContainText("Cheese");
    await expect(page.locator(".order-pane")).toContainText("₹80");
  });
});

test.describe("Admin browser flows", () => {
  test("adds a category and an item, filters the menu table, and exports invoices", async ({ page }) => {
    await page.goto("/admin");

    await expect(page.getByRole("heading", { name: "Menu Admin" })).toBeVisible();

    await page.getByRole("button", { name: "Manage Categories" }).click();
    const categoryModal = page.locator("#admin-manage-categories-modal");
    await expect(categoryModal).toBeVisible();
    await categoryModal.getByLabel("New Category").fill("UI Test Specials");
    await Promise.all([
      page.waitForResponse((response) => response.url().includes("/admin/categories/create") && response.status() === 303),
      categoryModal.getByRole("button", { name: "Add" }).click(),
    ]);

    await expect(page).toHaveURL(/\/admin$/);
    await page.getByRole("button", { name: "Add Item" }).click();
    const itemModal = page.locator("#admin-item-modal");
    await expect(itemModal).toBeVisible();
    await itemModal.getByLabel("Category").selectOption({ label: "UI Test Specials" });
    await itemModal.getByRole("textbox", { name: "Item", exact: true }).fill("Playwright Paneer");
    await itemModal.getByRole("textbox", { name: "Variant", exact: true }).fill("Regular");
    await itemModal.getByRole("spinbutton", { name: "Price", exact: true }).fill("321");
    await Promise.all([
      page.waitForResponse((response) => response.url().includes("/admin/items/create") && response.status() === 303),
      itemModal.getByRole("button", { name: "Add Item" }).click(),
    ]);

    await expect(page).toHaveURL(/\/admin$/);
    await expect(page.locator("input[name='name'][value='Playwright Paneer']")).toBeVisible();
    await page.getByLabel("Search").fill("Playwright Paneer");
    await expect(page.locator("[data-admin-menu-row]:visible")).toHaveCount(1);
    await expect(page.locator("[data-admin-menu-row]:visible input[name='name']")).toHaveValue("Playwright Paneer");

    await page.getByRole("button", { name: "Past Invoices" }).click();
    await expect(page.locator("#admin-invoices-modal")).toBeVisible();

    const downloadPromise = page.waitForEvent("download");
    await page.getByRole("link", { name: "Export CSV" }).click();
    const download = await downloadPromise;
    expect(download.suggestedFilename()).toBe("bill-of-fare-invoices.csv");
  });
});
