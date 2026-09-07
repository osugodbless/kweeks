import { describe, expect, it } from "vitest";
import { fmtNaira, naira, splitPodium } from "@/lib/player";
import { avatarById } from "@/lib/avatars";

describe("money formatting", () => {
  it("formats whole naira with thousands separators", () => {
    expect(fmtNaira("50000")).toBe("50,000");
    expect(fmtNaira("150000")).toBe("150,000");
    expect(fmtNaira(0)).toBe("0");
    expect(fmtNaira("1000000")).toBe("1,000,000");
  });

  it("handles empty/NaN input defensively", () => {
    expect(fmtNaira("")).toBe("0");
    expect(fmtNaira("abc")).toBe("0");
    expect(fmtNaira("12.5")).toBe("12");
  });

  it("prefixes the naira symbol", () => {
    expect(naira("50000")).toBe("₦50,000");
    expect(naira(25000)).toBe("₦25,000");
  });
});

describe("splitPodium", () => {
  it("mirrors the Go backend SplitPodium for the canonical 50k / 3 winner case", () => {
    // internal/domain/money.go: weights 3,2,1 over sum 6; rounding remainder to rank 1.
    expect(splitPodium("50000", 3)).toEqual([25001, 16666, 8333]);
  });

  it("gives everything to a single winner", () => {
    expect(splitPodium("50000", 1)).toEqual([50000]);
  });

  it("never overspends the pool", () => {
    for (const [pool, n] of [
      ["50000", 3],
      ["100000", 5],
      ["12345", 3],
      ["99999", 2],
    ] as const) {
      const shares = splitPodium(pool, n);
      expect(shares.reduce((a, b) => a + b, 0)).toBe(parseInt(pool, 10));
    }
  });

  it("guards invalid winner counts", () => {
    expect(splitPodium("50000", 0)).toEqual([]);
  });
});

describe("avatars", () => {
  it("resolves known ids and falls back for unknown", () => {
    expect(avatarById("ghost").id).toBe("ghost");
    expect(avatarById("cat").label).toBe("Cat");
    expect(avatarById("nope").id).toBe("ghost");
    expect(avatarById(undefined).id).toBe("ghost");
  });
});
