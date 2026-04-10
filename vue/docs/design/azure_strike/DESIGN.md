# Design System Document: Premium Gaming Editorial

## 1. Overview & Creative North Star
The Creative North Star for this system is **"The Kinetic Luminary."** 

Moving away from the static, blocky layouts typical of standard gaming interfaces, this system treats the UI as a fluid, high-energy editorial experience. We bridge the gap between "high-stakes excitement" and "premium sophistication." The design breaks the traditional template look by utilizing **intentional asymmetry**, where gaming cards and reward modules overlap their containers to create a sense of forward motion. We use high-density information architecture balanced by expansive white space to ensure the experience feels elite, not cluttered.

## 2. Colors
Our palette is a sophisticated blend of deep authority and vibrant energy. The core logic relies on a "Cold-to-Warm" transition: Blue provides the stable, premium foundation, while Red/Yellow highlights drive kinetic action.

### Color Tokens (Material Convention)
*   **Primary (Action/Brand):** `#004edb` (Vibrant Blue)
*   **Secondary (Rewards/Urgency):** `#b71211` (Deep Red)
*   **Surface:** `#f5f6fb` (Cool White-Blue)
*   **Surface Container Lowest:** `#ffffff` (Pure White for elevated cards)
*   **Tertiary (Wins/Highlights):** `#6c5a00` (Gold accents via `#ffd709` container)

### The "No-Line" Rule
**Strict Mandate:** Designers are prohibited from using 1px solid borders to section off the interface. Boundaries must be defined solely through background color shifts. For example, a `surface-container-low` section should sit directly on a `surface` background. The change in tonal value provides a cleaner, more modern separation that feels integrated rather than boxed in.

### The Glass & Gradient Rule
To achieve a "Premium Gaming" feel, floating elements (Modals, Navigation Bars, Hovering CTAs) must utilize **Glassmorphism**. Use semi-transparent surface colors with a `24px` backdrop-blur. 
*   **Signature Textures:** Main CTAs should never be flat. Use a subtle linear gradient (Top-Left to Bottom-Right) transitioning from `primary` (`#004edb`) to `primary-container` (`#7e9cff`) to add "soul" and depth to the interaction.

---

## 3. Typography
We use **Plus Jakarta Sans** for its modern, geometric clarity and "tech-forward" personality.

*   **Display (Lg/Md):** Used for jackpot totals and major wins. Letter spacing set to `-0.02em` for a tighter, high-end editorial feel.
*   **Headline (Sm-Lg):** Used for game category titles. These drive the brand's authoritative voice.
*   **Title (Md/Sm):** Used for card titles. Bold weight (`700`) is mandatory to maintain hierarchy in high-density layouts.
*   **Body (Md):** The workhorse for descriptions. We use a slightly increased line-height (`1.6`) to ensure readability against vibrant backgrounds.
*   **Label (Sm):** Used for micro-copy and metadata. Always in Medium weight to ensure the ink doesn't "bleed" on high-res mobile displays.

---

## 4. Elevation & Depth
In this system, depth is a product of **Tonal Layering** rather than structural lines.

*   **The Layering Principle:** Treat the UI as stacked sheets of fine paper. 
    *   *Level 0:* `surface` (The canvas).
    *   *Level 1:* `surface-container-low` (Content groupings).
    *   *Level 2:* `surface-container-lowest` (The interactive card itself).
*   **Ambient Shadows:** For floating elements like the "Add to Home Screen" prompt or Reward Popups, use an extra-diffused shadow: `0px 12px 32px rgba(0, 78, 219, 0.08)`. Note the use of a blue-tinted shadow instead of grey to maintain color harmony.
*   **The "Ghost Border" Fallback:** If a border is required for accessibility, use the `outline-variant` token at **15% opacity**. It should be felt, not seen.

---

## 5. Components

### Buttons (Kinetic CTAs)
*   **Primary:** Gradient-filled (`primary` to `primary-container`), `xl` (1.5rem) rounded corners. Use `on-primary` for text.
*   **Secondary (Reward):** Solid `secondary` (`#b71211`) with a subtle `secondary_dim` outer glow to signal high-value interaction.
*   **Tertiary (Ghost):** No background, `primary` text, bold weight.

### Cards & Game Modules
*   **Construction:** Use `surface-container-lowest` backgrounds with `lg` (1rem) corner radius.
*   **Spacing:** Forbid the use of divider lines. Separate "Winning Information" list items using `12px` of vertical white space and subtle alternating backgrounds (`surface-container-low` vs `surface-container-lowest`).
*   **Gaming Density:** Use `title-sm` for game names to allow for higher grid density (2 or 3 columns) without sacrificing legibility.

### Floating Bottom Navigation
*   **Style:** Glassmorphic container with a `20px` blur. 
*   **Active State:** The active icon should sit within a `primary-container` circular "halo" to provide immediate visual feedback.

### Rewards/Promo Chips
*   Vibrant, high-contrast pills using `secondary_container` with `on_secondary_container` text. These should appear "over-inflated" with `full` roundedness to look like premium physical tokens.

---

## 6. Do's and Don'ts

### Do
*   **DO** use overlapping elements. Let a game icon "bleed" outside its container by 8-10px to create depth.
*   **DO** use the `tertiary` (Gold/Yellow) sparingly for "VIP" or "Super Win" moments only.
*   **DO** prioritize high-density information in list views while maintaining a minimum of `16px` outer-screen padding.

### Don't
*   **DON'T** use black (#000000) for shadows. Always tint shadows with the primary blue or secondary red to keep the "Premium Gaming" glow.
*   **DON'T** use 90-degree corners. Everything must have a minimum of `sm` (0.25rem) roundedness to feel approachable.
*   **DON'T** use "divider lines" between list items. Use tonal shifts in the background to indicate a new row.
*   **DON'T** use flat, matte colors for major CTAs. Always include a 5-10% vertical gradient to simulate a premium, tactile surface.