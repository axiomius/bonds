import { test, expect } from '@playwright/test';

// Verifies the end-to-end loop that was broken before the fix series:
// the LanguageSwitcher widget actually flips the rendered UI text, and
// switching to a non-English language before registering causes the
// backend to seed the vault in that language (proving the Accept-Language
// interceptor is wired through to the seed code).
test.describe('i18n end-to-end', () => {
  test('language switcher flips rendered chrome on the register page', async ({ page }) => {
    await page.goto('/register');

    // English baseline: the submit button reads "Create account".
    await expect(page.getByRole('button', { name: /create account/i })).toBeVisible();

    // Open the language switcher (aria-label="language") and pick 中文.
    await page.getByRole('button', { name: 'language' }).first().click();
    await page.getByRole('menuitem', { name: '中文' }).click();

    // Same submit button now reads in Chinese — that's i18n applied live
    // without a reload.
    await expect(page.getByRole('button', { name: '创建账户' })).toBeVisible({
      timeout: 5000,
    });
  });

  test('registering with Chinese UI persists locale across reload', async ({ page }) => {
    await page.goto('/register');

    // Switch to Chinese FIRST so the axios interceptor sends
    // Accept-Language=zh on the /register call, which is what feeds the
    // seed routines.
    await page.getByRole('button', { name: 'language' }).first().click();
    await page.getByRole('menuitem', { name: '中文' }).click();
    await expect(page.getByRole('button', { name: '创建账户' })).toBeVisible({
      timeout: 5000,
    });

    // Fill the form by label, not by placeholder, because placeholders
    // localize and would shuffle the .first() ordering.
    const email = `i18n-zh-${Date.now()}@example.com`;
    const textInputs = page.locator('input[type="text"], input:not([type])');
    await textInputs.nth(0).fill('测试');
    await textInputs.nth(1).fill('用户');
    await page.locator('input[type="email"], input[placeholder*="邮箱"]').first().fill(email);
    await page.locator('input[type="password"]').first().fill('password123');
    await page.getByRole('button', { name: '创建账户' }).click();

    await expect(page).toHaveURL(/\/vaults/, { timeout: 15000 });

    // The /vaults page chrome must render in Chinese — proving the
    // Accept-Language header was sent (so the React app loaded the zh
    // bundle) and that LanguageSwitcher's change persisted through the
    // POST /register round-trip. "保险库" is the Chinese for "Vault" and
    // is present in the empty-state heading and the create button.
    await expect(page.getByRole('heading', { name: '保险库' })).toBeVisible();
    await expect(page.getByRole('button', { name: /新建保险库|创建保险库/ }).first()).toBeVisible();

    // Reload to confirm the locale isn't just session-scoped: the saved
    // user.locale on the server should drive usePreferencesSync to apply
    // zh on every subsequent boot.
    await page.reload();
    await expect(page.getByRole('heading', { name: '保险库' })).toBeVisible({
      timeout: 10000,
    });
  });
});
