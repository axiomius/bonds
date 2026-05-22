# Bonds Frontend

Bonds uses a modern React single-page application frontend. It compiles to static files that the Go backend embeds directly in its single-binary distribution.

## Tech Stack

- **Framework**: React 19, Vite 7
- **Language**: TypeScript (Strict Mode)
- **UI Library**: Ant Design v6
- **Data Fetching**: TanStack Query v5
- **Routing**: React Router v7
- **HTTP Client**: Axios
- **Internationalization**: react-i18next

## Directory Structure

```text
web/
  src/
    api/                    # API client layer
      generated/            # Auto-generated API client modules (gitignored)
      index.ts              # Entry point establishing Axios instance and handlers
    components/             # Shared reusable UI elements
    locales/                # Localization JSON dictionaries (en.json, zh.json)
    pages/                  # Routed pages categorized by domain
    stores/                 # Global contexts (Auth, Theme)
    types/                  # TypeScript interface overrides and custom declarations
    utils/                  # Date formatters, helper functions, and shared types
    test/                   # Vitest unit test environment and mocks
  e2e/                      # Playwright E2E test suites
```

## Running Development Server

To build the client, you first need to run the API generation pipeline. The frontend code depends on type declarations generated from the backend OpenAPI schema.

Ensure you use `bun` instead of `npm` or `yarn`.

```bash
# Install dependencies from root
make setup

# Generate the API client
make gen-api

# Start the frontend dev server
cd web
bun run dev
```

The Vite server runs at `http://localhost:5173`. It automatically proxies API requests to the Go backend running on port `8080`.

## API Client Architecture

Do not edit files inside `src/api/generated/` directly. They are completely overwritten when the generator runs.

The generation flow works like this:
1. Go handler decorators define endpoints and DTO shapes.
2. `make swagger` compiles these into `server/docs/swagger.json`.
3. `make gen-api` runs `swagger-typescript-api` using the compiled schema.
4. The output is written to `web/src/api/generated/`.

To call an endpoint, import the global API instance:
```typescript
import { api } from "@/api";

// Call directly
const response = await api.contacts.contactsList({ vaultId });
```

## Styling and Themes

Bonds relies on Ant Design v6 token-based styling. The theme configuration resides in `src/stores/ThemeProvider.tsx`. It supports:
- **Light mode**
- **Dark mode**
- **System preference** (automatically syncs with system settings)

Use standard CSS classes or Ant Design styled wrappers where custom offsets are necessary. Avoid writing custom stylesheets.

## Internationalization (i18n)

Every user-visible string must use the translation hook:
```typescript
import { useTranslation } from "react-i18next";

const { t } = useTranslation();
return <p>{t("vault.dashboard.title")}</p>;
```

Add your translation keys in both `web/src/locales/en.json` and `web/src/locales/zh.json`. The CI suite checks for key parity, and any missing keys will fail the build.

## Frontend Development Best Practices

### Modal and Edit Forms

Ant Design form instances hold state across renders. A common pitfall is editing a record in a modal, closing it, opening another record, and seeing the old values.

To avoid this behavior, follow these patterns:
- Extract the Form and its `Form.useForm()` hook into a distinct inner child component.
- Keep only the Modal wrapper in the parent component.
- Render the child component with a key tied to the active record ID: `<EditFormInner key={record.id} />`.
- When the active record changes, React unmounts the old form and mounts a new one. This ensures the hook state resets correctly.
- Add form buttons inside the form component and set the Modal `footer` to `null`.

### Component State

Prefer controlled components where a parent manages state through props. Avoid syncing props with local state using `useEffect`, which leads to synchronization bugs.

### Date Formatting

Always format dates using the `useDateFormat()` hook from `src/utils/dateFormat.ts`. This ensures users see dates in their chosen locales. Avoid hardcoding custom format strings.

## Testing

### Unit and Integration Tests

Bonds uses Vitest for unit testing.
```bash
# Run all unit tests
bun run test

# Run a specific test file
bun run test -- src/test/Login.test.tsx
```

All network calls must be mocked to run tests reliably. Components making direct API requests need their HTTP client mocked in `src/test/setup.ts` or inside the test block.

### End-to-End (E2E) Tests

Playwright drives end-to-end user flows.
```bash
# Run all E2E tests
bunx playwright test

# Run a specific E2E test file
bunx playwright test e2e/auth.spec.ts
```

The E2E suite starts its own Go backend server on an ephemeral port. It automatically cleans the testing database before each execution.
