import { describe, expect, it } from "vitest";
import { getReadableLabelTagColors } from "@/utils/labelColor";

describe("getReadableLabelTagColors", () => {
  it("uses black text when configured white text is unreadable on a light background", () => {
    const colors = getReadableLabelTagColors("#fff8b0", "#ffffff");

    expect(colors.style.backgroundColor).toBe("#fff8b0");
    expect(colors.style.color).toBe("#000000");
    expect(colors.style.borderColor).toBe("rgba(0, 0, 0, 0.18)");
  });

  it("uses white text when configured black text is unreadable on a dark background", () => {
    const colors = getReadableLabelTagColors("#111827", "#000000");

    expect(colors.style.backgroundColor).toBe("#111827");
    expect(colors.style.color).toBe("#ffffff");
  });

  it("keeps already readable configured text colors", () => {
    const colors = getReadableLabelTagColors("#111827", "#f8fafc");

    expect(colors.style.color).toBe("#f8fafc");
  });

  it("normalizes shorthand hex colors", () => {
    const colors = getReadableLabelTagColors("#fff", "#fff");

    expect(colors.style.backgroundColor).toBe("#ffffff");
    expect(colors.style.color).toBe("#000000");
  });

  it("supports named Ant Design tag colors", () => {
    const colors = getReadableLabelTagColors("green", "#ffffff");

    expect(colors.style.backgroundColor).toBe("#f6ffed");
    expect(colors.style.color).toBe("#000000");
  });

  it("falls back to the default tag style for unknown background colors", () => {
    const colors = getReadableLabelTagColors("not-a-color", "#ffffff");

    expect(colors.color).toBe("default");
    expect(colors.style).toEqual({});
  });
});
