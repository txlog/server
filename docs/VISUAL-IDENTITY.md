# **Txlog: Visual Identity & UI Style Guide**

## **1\. Brand Overview**

**Txlog** represents organization, creativity, and seamless tracking. Inspired
by the visual of a well-kept notebook and pencil, the brand identity is designed
to feel approachable, friendly, and highly structured.

* **Brand Personality:** Friendly, Reliable, Organized, Creative, Accessible.
* **Design Style:** Flat design, soft geometry, highly rounded corners, clean
  and modern.

## **2\. Color Palette**

The color palette is directly extracted from the brand's primary iconography. It
balances a grounded dark tone with a soft background and vibrant accents used
for categorization and highlights. All text colors comply with **WCAG AA**
contrast requirements (minimum 4.5:1 ratio on white backgrounds).

### **Primary Colors**

These form the foundation of the Txlog UI.

* **Indigo Slate (Primary Dark):** \#424565
  * *Usage:* Text, headers, main navigation, sidebar backgrounds, tooltips,
    and high-emphasis borders (inspired by the notebook spine and pencil lead).
  * *Contrast:* ~10.5:1 on white ✅
* **Lavender Frost (Primary Light):** \#E6E6FA
  * *Usage:* Borders, separators, table dividers, subtle interactive states,
    and muted background highlights (inspired by the notebook cover).
* **Soft Background:** \#F8F9FE
  * *Usage:* App body background. A near-white with a subtle blue tint that
    harmonizes with the indigo text.
* **White (Base Surface):** \#FFFFFF
  * *Usage:* Primary container backgrounds (cards, modals) to ensure high
    contrast and readability.

### **Accent & Semantic Colors**

Used for actions, status indicators, and categorization tags.

* **Coral Bookmark (Action/Destructive):** \#D9556A
  * *Usage:* Destructive actions (delete, deactivate), error states, warning
    bars on modals, and important highlights.
  * *Contrast:* ~4.6:1 on white ✅
* **Golden Label (Warning/Highlight):** \#F4B54B
  * *Usage:* Warning alerts, "needs restart" indicators, badges, and the
    Reinstall action icon. **Not used as text on white** due to low contrast;
    always paired with a colored background.
  * *Contrast:* ~1.9:1 on white ⚠️ (use on backgrounds only)
* **Sky Blue (Info/Active):** \#4A8AE8
  * *Usage:* Primary call-to-action (CTA) buttons, active states, text links,
    search buttons, informational banners, and progress bars.
  * *Contrast:* ~4.6:1 on white ✅
* **Leaf Green (Success):** \#4A9E42
  * *Usage:* Success messages, Install/Upgrade action icons, "active" status
    dots, and positive feedback modals.
  * *Contrast:* ~4.5:1 on white ✅
* **Purple (Metadata/Tertiary):** \#8B5CF6
  * *Usage:* Obsolete and Reason Change action icons, Administrator role badge.
    Provides visual distinction from the main semantic colors.
  * *Contrast:* ~4.6:1 on white ✅

## **3\. Typography**

To match the soft, rounded aesthetic of the logo, the typography must be clean,
legible, and slightly geometric with friendly curves.

* **Primary Font (Headings & Titles):** *Poppins*
  * *Weight:* Semi-Bold (600) to Bold (700)
  * *Characteristics:* Round, welcoming, excellent for large display text.
* **Secondary Font (Body Text & UI):** *Inter*
  * *Weight:* Regular (400) to Medium (500)
  * *Characteristics:* Highly legible at small sizes, neutral, balances the
    playfulness of the primary font.

## **4\. Iconography & Shapes**

The Txlog visual language relies heavily on its shapes.

* **Icon Library:** [Lucide Icons](https://lucide.dev) — a modern, stroke-based
  icon set with 1500+ icons. Used via CDN with `data-lucide` attributes and
  initialized with `lucide.createIcons()`.
* **Border Radius:** \* *Global UI:* All cards, modals, and buttons must use a
  large border radius to match the notebook icon.
  * *Value:* 12px to 16px for standard elements (buttons, inputs). 24px for
    larger layout containers or cards.
  * **NO sharp edges.** Everything should feel smooth and safe.
* **Icons:** \* Use solid, flat-style icons or thick stroke icons (min. 2px
  width) with rounded caps and joints.
  * Keep details minimal to match the simplicity of the Txlog logo.

## **5\. UI Elements**

### **Buttons**

* **Primary Button:** \* Background: Coral Bookmark (\#D9556A) or Sky Blue
  (\#4A8AE8).
  * Text Color: White.
  * Border Radius: 12px.
  * Hover State: Slight upward lift transform (translateY(-2px)) with a
    colored shadow glow.
* **Secondary Button:**
  * Background: Transparent.
  * Border: 2px solid Lavender Frost (\#E6E6FA) or Indigo Slate (\#424565).
  * Text Color: Indigo Slate (\#424565).
  * Border Radius: 12px.

### **Input Fields**

* Background: White or very light gray.
* Border: 2px solid Lavender Frost (\#E6E6FA).
* Focus State: Border changes to Sky Blue (\#4A8AE8) with a soft glow
  (box-shadow).
* Border Radius: 12px.

### **Cards & Containers**

* Background: White.
* Border Radius: 16px or 24px.
* Shadow: Soft, diffused drop shadow to create depth without making it look
  heavy.
  * *CSS Example:* box-shadow: 0 8px 24px rgba(66, 69, 101, 0.08);

### **Modals**

* Border Radius: 24px (rounded-3xl).
* Use `overflow-hidden` to clip top colored status bars.
* Destructive modals include a thin Coral bar at the top.
* Success modals include a thin Leaf Green bar at the top.
* Animated entrance: scale-95 → scale-100 with opacity transition.

### **Tooltips**

* Background: Indigo Slate (\#424565).
* Text: White, text-xs.
* Appear on hover via `group-hover:block` pattern.
* Positioned above the element with centered arrow alignment.

### **Status Dots**

* Green (animated pulse): Asset seen in the last 24 hours.
* Golden (static): Asset seen between 1–15 days ago.
* Coral (static): Asset not seen for 15+ days.
