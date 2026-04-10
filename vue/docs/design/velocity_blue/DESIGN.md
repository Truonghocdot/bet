# Design System Strategy: High-Velocity Luxury

## 1. Overview & Creative North Star
The Creative North Star for this design system is **"The Kinetic Architect."** 

In the high-stakes world of sports betting and gaming, "modern" and "clean" are no longer enough. To feel premium, the UI must balance high-energy momentum with an authoritative, architectural stillness. We break the "generic template" look by utilizing intentional asymmetry—where content isn't just center-aligned, but dynamically positioned to lead the eye—and by employing a sophisticated depth model that moves away from flat, boxed-in layouts.

This system is built to feel like a high-end digital lounge: expansive, responsive, and vibrating with potential, while maintaining the absolute clarity required for complex data.

---

## 2. Colors: Tonal Depth & Signature Polish
We are moving beyond "Blue and White" into a spectrum of atmospheric depths. 

### The Palette
- **Primary High-Energy:** `primary` (#0058bb) serves as our brand anchor. Use `primary_container` (#6c9fff) for interactive surfaces to create a "glowing" effect.
- **The Neutrals:** `surface` (#f7f5ff) is our canvas. It is intentionally cool-toned to maintain a "trustworthy" and clinical precision.
- **Accents:** `tertiary_container` (#fdd404) is used sparingly for "Big Win" moments and urgent CTAs to provide high-contrast energy against the blue.

### Strategic Color Rules
*   **The "No-Line" Rule:** 1px solid borders are strictly prohibited for sectioning. We define space through color shifts. A `surface_container_low` section sitting on a `surface` background provides all the separation necessary.
*   **Surface Hierarchy & Nesting:** Think of the UI as stacked sheets of glass. 
    *   *Base:* `surface`
    *   *Section:* `surface_container`
    *   *Card:* `surface_container_lowest` (Pure White #ffffff) to create a natural "pop" against the cool background.
*   **The Glass & Gradient Rule:** For floating navigation or promotional modals, use Glassmorphism: `surface_container_lowest` at 80% opacity with a `24px` backdrop-blur. 
*   **Signature Textures:** Use a linear gradient from `primary` to `primary_dim` at a 135-degree angle for hero banners and main action buttons to add a "liquid metal" premium feel.

---

## 3. Typography: Editorial Authority
We utilize a dual-font system to separate "Action" from "Information."

*   **The Headline Voice (Plus Jakarta Sans):** Used for `display` and `headline` roles. This typeface features modern, wide apertures that feel premium and expansive. Use `display-lg` (3.5rem) with tight letter-spacing (-0.02em) for promotional headlines to create a sense of scale.
*   **The Functional Voice (Inter):** Used for `title`, `body`, and `label`. Inter provides maximum legibility for betting odds and data tables. 
*   **Hierarchical Contrast:** Pair a `headline-sm` in bold with a `body-md` in `on_surface_variant` to create clear editorial lanes. Typography should never feel "standard"; use the `label-sm` (uppercase, 0.05em tracking) for category tags to mimic high-end magazine layouts.

---

## 4. Elevation & Depth: Tonal Layering
We move away from the "shadow-heavy" look of 2010s apps toward **Tonal Layering.**

*   **The Layering Principle:** Depth is achieved by placing a `surface_container_lowest` card on top of a `surface_container_high` background. The contrast in brightness creates the lift.
*   **Ambient Shadows:** If a card must float (e.g., a live bet slip), use an ultra-diffused shadow: `Y: 12px, Blur: 40px, Color: on_surface @ 6%`. It should feel like a soft glow, not a dark smudge.
*   **The "Ghost Border" Fallback:** If a container needs more definition, use `outline_variant` at 15% opacity. This creates a "hairline" effect that feels like etched glass rather than a heavy stroke.
*   **Asymmetric Breathing Room:** Do not pad elements equally on all sides. Use the spacing scale to give more "headroom" (top padding) to sections, creating an airy, luxury-retail feel.

---

## 5. Components: Functional Elegance

### Buttons & Interaction
*   **Primary Action:** `primary` background with `on_primary` text. Use `xl` (1.5rem) roundedness for a friendly but modern feel. Add a subtle inner-glow (top-down white gradient at 10% opacity) for a 3D "pressable" effect.
*   **Secondary/Betting Odds:** Use `surface_container_high`. Upon hover, transition to `primary_container`. No borders.

### Game Cards & Promotions
*   **Game Cards:** Use `surface_container_lowest`. Forbid dividers. Content is separated by `body-sm` typography and `0.5rem` vertical gaps.
*   **Promotional Banners:** Use "The Kinetic Architecture" layout. The subject (e.g., an athlete or dealer) should "break the box" and overlap the top or side edge of the container to create depth.
*   **Navigation Menus:** The bottom nav should use the Glassmorphism rule. Active states are indicated by a `primary` color shift in the icon and a `label-md` bolding—no bulky background chips.

### Input & Form Fields
*   **Text Inputs:** Use `surface_container_low` with a `none` border. On focus, a `2px` `primary` "Ghost Border" (20% opacity) appears. This maintains a clean look while providing clear feedback.

---

## 6. Do's and Don'ts

### Do
*   **Do** use extreme white space. If you think there is enough space, add 8px more.
*   **Do** use `plusJakartaSans` for large numbers (odds/payouts). It is the "hero" of the data.
*   **Do** nest containers (Lowest on Low) to create a sense of organized complexity.
*   **Do** use subtle motion—elements should "slide" into place with a cubic-bezier(0.16, 1, 0.3, 1) easing.

### Don't
*   **Don't** use pure black (#000000) for text. Use `on_surface` (#232c51) for a sophisticated, deep-navy professional tone.
*   **Don't** use 1px solid borders to separate list items. Use a 12px vertical gap or a subtle background shift.
*   **Don't** use harsh drop shadows. If it looks like a shadow, it’s too heavy. It should look like "ambient occlusion."
*   **Don't** crowd the "Big Win" or promotional areas. Let the imagery speak through the Glassmorphism layers.