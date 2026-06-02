import { describe, it, expect, vi, beforeAll, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes, useLocation } from "react-router-dom";
import { App as AntApp, ConfigProvider } from "antd";
import ContactList from "@/pages/contact/ContactList";
import type { Contact, PaginationMeta } from "@/api";

beforeAll(() => {
  globalThis.ResizeObserver = class {
    observe() {}
    unobserve() {}
    disconnect() {}
  };
});

function LocationProbe() {
  const location = useLocation();
  return <div data-testid="location-probe">{location.pathname}{location.search}</div>;
}

vi.mock("@/api", () => ({
  api: {
    contacts: {
      contactsList: vi.fn(),
      contactsLabelsDetail: vi.fn(),
      contactsSortUpdate: vi.fn(),
    },
    contactLabels: { contactLabelsList: vi.fn() },
    vaultSettings: { settingsLabelsList: vi.fn() },
    vcard: { contactsExportList: vi.fn(), contactsImportCreate: vi.fn() },
  },
  httpClient: {
    instance: {
      get: vi.fn().mockRejectedValue(new Error("mocked")),
    },
  },
}));

vi.mock("@/components/ContactAvatar", () => ({
  default: () => <div data-testid="contact-avatar" />,
}));

const mockUseQuery = vi.fn();
vi.mock("@tanstack/react-query", () => ({
  useQuery: (...args: unknown[]) => mockUseQuery(...args),
  useMutation: () => ({ mutate: vi.fn(), isPending: false }),
  useQueryClient: () => ({ invalidateQueries: vi.fn() }),
}));

function mockContactListQuery(contacts: Contact[] = [], meta: PaginationMeta = { total: contacts.length }) {
  mockUseQuery.mockImplementation((opts) => {
    const key = Array.isArray(opts?.queryKey) ? opts.queryKey : [];
    if (key.includes("labels")) {
      return { data: [], isLoading: false };
    }
    if (key[0] === "vaults" && key[2] === "contacts") {
      return { data: { contacts, meta }, isLoading: false };
    }
    return { data: undefined, isLoading: false };
  });
}

function getContactsQueryKey() {
  const call = mockUseQuery.mock.calls.find(([opts]) => {
    const key = Array.isArray(opts?.queryKey) ? opts.queryKey : [];
    return key[0] === "vaults" && key[2] === "contacts";
  });
  return call?.[0]?.queryKey as unknown[] | undefined;
}

function renderContactList(initialUrl = "/vaults/1/contacts") {
  return render(
    <ConfigProvider>
      <AntApp>
        <MemoryRouter initialEntries={[initialUrl]}>
          <Routes>
            <Route path="/vaults/:id/contacts" element={
              <>
                <ContactList />
                <LocationProbe />
              </>
            } />
            <Route path="/vaults/:id/contacts/:contactId" element={
              <>
                <LocationProbe />
              </>
            } />
          </Routes>
        </MemoryRouter>
      </AntApp>
    </ConfigProvider>,
  );
}

describe("ContactList", () => {
  beforeEach(() => {
    mockUseQuery.mockReset();
    mockContactListQuery();
  });

  it("renders loading state", () => {
    mockUseQuery.mockReturnValue({ data: undefined, isLoading: true });
    renderContactList();
    expect(document.querySelector(".ant-spin")).toBeInTheDocument();
  }, 15000);

  it("renders empty state", () => {
    mockContactListQuery();
    renderContactList();
    expect(screen.getByText("No contacts yet")).toBeInTheDocument();
  });

  it("renders search input", () => {
    mockContactListQuery();
    renderContactList();
    expect(
      screen.getByPlaceholderText("Quick search"),
    ).toBeInTheDocument();
  });

  it("reads page and per_page from URL query parameters", () => {
    renderContactList("/vaults/1/contacts?page=3&per_page=50");

    expect(getContactsQueryKey()).toEqual([
      "vaults",
      "1",
      "contacts",
      null,
      3,
      50,
      "name",
      "",
      "active",
    ]);
  });

  it("falls back to default pagination when URL query values are invalid", () => {
    renderContactList("/vaults/1/contacts?page=abc&per_page=0");

    expect(getContactsQueryKey()).toEqual([
      "vaults",
      "1",
      "contacts",
      null,
      1,
      20,
      "name",
      "",
      "active",
    ]);
  });

  it("updates URL when pagination changes", async () => {
    const user = userEvent.setup();
    mockContactListQuery(
      Array.from({ length: 20 }).map((_, i) => ({
        id: String(i + 1),
        first_name: `User ${i + 1}`,
        last_name: "Example",
        updated_at: "2024-06-01T00:00:00Z",
      })),
      { total: 60 },
    );

    renderContactList("/vaults/1/contacts");

    const page2Button = document.querySelector<HTMLElement>(".ant-pagination-item-2 a");
    expect(page2Button).toBeInTheDocument();
    if (!page2Button) throw new Error("Page 2 pagination link was not rendered");
    await user.click(page2Button);

    await waitFor(() => {
      expect(screen.getByTestId("location-probe")).toHaveTextContent("/vaults/1/contacts?page=2&per_page=20");
    });
  });

  it("preserves pagination query parameters when navigating to a contact", async () => {
    const user = userEvent.setup();
    mockContactListQuery(
      [{
        id: "42",
        first_name: "Test",
        last_name: "User",
        updated_at: "2024-06-01T00:00:00Z",
      }],
      { total: 100 },
    );

    renderContactList("/vaults/1/contacts?page=3&per_page=50");

    const contactRow = await screen.findByText("Test User");
    await user.click(contactRow);

    await waitFor(() => {
      expect(screen.getByTestId("location-probe")).toHaveTextContent("/vaults/1/contacts/42?page=3&per_page=50");
    });
  });
});
