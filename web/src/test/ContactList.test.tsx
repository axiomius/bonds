import { describe, it, expect, vi, beforeAll, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes, useLocation } from "react-router-dom";
import { App as AntApp, ConfigProvider } from "antd";
import ContactList from "@/pages/contact/ContactList";

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
    mockUseQuery.mockReturnValue({ data: undefined, isLoading: false });
  });

  it("renders loading state", () => {
    mockUseQuery.mockReturnValue({ data: undefined, isLoading: true });
    renderContactList();
    expect(document.querySelector(".ant-spin")).toBeInTheDocument();
  }, 15000);

  it("renders empty state", () => {
    mockUseQuery.mockImplementation((opts) => {
      if (Array.isArray(opts?.queryKey) && opts.queryKey.includes("labels")) {
        return { data: [], isLoading: false };
      }
      return { data: { data: [], meta: { pagination: { current_page: 1, last_page: 1, total: 0, per_page: 20 } } }, isLoading: false };
    });
    renderContactList();
    expect(screen.getByText("No contacts yet")).toBeInTheDocument();
  });

  it("renders search input", () => {
    mockUseQuery.mockImplementation((opts) => {
      if (Array.isArray(opts?.queryKey) && opts.queryKey.includes("labels")) {
        return { data: [], isLoading: false };
      }
      return { data: { data: [], meta: { pagination: { current_page: 1, last_page: 1, total: 0, per_page: 20 } } }, isLoading: false };
    });
    renderContactList();
    expect(
      screen.getByPlaceholderText("Quick search"),
    ).toBeInTheDocument();
  });

  it("reads page and per_page from URL query parameters", () => {
    renderContactList("/vaults/1/contacts?page=3&per_page=50");
    
    expect(mockUseQuery).toHaveBeenCalledWith(
      expect.objectContaining({
        queryKey: expect.arrayContaining([
          expect.objectContaining({ page: 3, per_page: 50 })
        ]),
      })
    );
  });

  it("updates URL when pagination changes", async () => {
    const user = userEvent.setup();
    mockUseQuery.mockImplementation((opts) => {
      if (Array.isArray(opts?.queryKey) && opts.queryKey.includes("labels")) {
        return { data: [], isLoading: false };
      }
      return { 
        data: { 
          data: Array.from({ length: 20 }).map((_, i) => ({ id: i + 1, first_name: `User ${i}` })), 
          meta: { pagination: { current_page: 1, last_page: 3, total: 60, per_page: 20 } } 
        }, 
        isLoading: false 
      };
    });
    
    renderContactList("/vaults/1/contacts");
    
    const page2Button = await screen.findByRole("button", { name: "2" });
    await user.click(page2Button);
    
    await waitFor(() => {
      expect(screen.getByTestId("location-probe")).toHaveTextContent("/vaults/1/contacts?page=2");
    });
  });

  it("preserves pagination query parameters when navigating to a contact", async () => {
    const user = userEvent.setup();
    mockUseQuery.mockImplementation((opts) => {
      if (Array.isArray(opts?.queryKey) && opts.queryKey.includes("labels")) {
        return { data: [], isLoading: false };
      }
      return { 
        data: { 
          data: [{ id: 42, first_name: "Test User" }], 
          meta: { pagination: { current_page: 3, last_page: 5, total: 100, per_page: 50 } } 
        }, 
        isLoading: false 
      };
    });
    
    renderContactList("/vaults/1/contacts?page=3&per_page=50");
    
    const contactRow = await screen.findByText("Test User");
    await user.click(contactRow);
    
    await waitFor(() => {
      expect(screen.getByTestId("location-probe")).toHaveTextContent("/vaults/1/contacts/42?page=3&per_page=50");
    });
  });
});
