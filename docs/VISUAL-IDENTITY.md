# **Txlog: Visual Identity & UI Style Guide**

## **1. Brand Overview**

**Txlog** represents organization, seamless tracking, and infrastructure
management. Following the adoption of the Kumo UI design system, the brand
identity has evolved to a neutral, corporate, and highly structured
aesthetic—inspired by modern dashboards.

- **Brand Personality:** Professional, Reliable, Organized, Neutral, Accessible.
- **Design Style:** Clean layout, semantic token-based styling, subtle borders,
  data-dense but readable.

## **2. Color Palette (Kumo UI Tokens)**

The color palette is built entirely on **Kumo UI semantic tokens**. We do not
use hardcoded hex values or arbitrary Tailwind color classes; everything maps to
`--kumo-*` variables (e.g., `text-kumo-default`, `bg-kumo-canvas`).

### **Surfaces & Backgrounds**

- **Canvas (`bg-kumo-canvas`):** The base background of the application. Light
  and neutral.
- **Control (`bg-kumo-control`):** Background for interactive elements, cards,
  and primary containers on top of canvas.
- **Tint (`bg-kumo-tint`):** Subtle background for hovers, selected states, or
  secondary container highlights.

### **Text Colors**

- **Default (`text-kumo-default`):** Primary text color. Used for headings, body
  text, and critical information.
- **Subtle (`text-kumo-subtle`):** Secondary text color. Used for descriptions,
  table headers, and less emphasized data.
- **Muted (`text-kumo-muted`):** Tertiary text color. Used for placeholders,
  disabled states, and very minor metadata.

### **Borders**

- **Line (`border-kumo-line`):** Standard borders for cards, inputs, and
  structural separation.
- **Hairline (`border-kumo-hairline`):** Very subtle dividers (like table rows
  or navbar bottom border).

### **Semantic & Accent Colors**

Used for actions, status indicators, and alerts.

- **Brand (`kumo-brand`):** Primary actions, text links, active tabs, and main
  CTAs.
- **Danger (`kumo-danger`):** Errors, destructive actions (Delete), and critical
  vulnerabilities.
- **Warning (`kumo-warning`):** Warnings, pending states, and medium/high
  severity indicators.
- **Success (`kumo-success`):** Success messages, positive trends, and "Fixed"
  statuses.

## **3. Typography**

To match the neutral, data-heavy aesthetic, the typography is highly legible and
standardized.

- **Primary & Secondary Font:** _Inter_
- **Weights:**
  - Regular (400) for body and secondary text (`font-normal`).
  - Medium (500) for most UI controls, buttons, and subtle emphasis
    (`font-medium`).
  - Semi-Bold (600) and Bold (700) for headings and important numeric data
    (`font-semibold`, `font-bold`).
- **Note:** The previous `Poppins` display font has been completely removed in
  favor of a unified `Inter` experience.

## **4. Iconography**

- **Icon Library:** [Phosphor Icons](https://phosphoricons.com/)
- **Style:** Outlined/Vazado (`weight="regular"`).
- **Implementation:** Inline SVG. We do not use CDNs or JS libraries to render
  icons.
- **Stroke Width:** `stroke-width="16"` on a `viewBox="0 0 256 256"`.
- **Note:** The previous `Lucide` icon library has been entirely removed.

## **5. UI Elements**

### **Buttons**

All buttons use the `data-kumo-component="Button"` styling pattern.

- **Primary Button:** `bg-kumo-brand` text `white`, rounded-lg.
- **Ghost/Secondary Button:** Transparent background, `text-kumo-brand` or
  `text-kumo-danger`, `hover:bg-kumo-tint`.
- **Focus:** Must have `focus-visible:ring-2 focus-visible:ring-kumo-brand`.

### **Inputs & Selects**

All form elements use `data-kumo-component="Input"` or
`data-kumo-component="Select"`.

- **Background:** `bg-kumo-canvas` or `bg-kumo-control`.
- **Border:** `border-kumo-line` with `rounded-lg`.
- **Text:** `text-kumo-default` (sm or base size).
- **Focus:** `focus:border-kumo-brand focus:ring-2 focus:ring-kumo-brand/20`.

### **Cards & Containers**

- **Background:** `bg-kumo-control`
- **Border:** `border-kumo-line`
- **Border Radius:** `rounded-xl` or `rounded-lg`.
- **Shadow:** `shadow-sm` (minimal depth).

### **Tables**

- Tables are structured using the `<table class="kumo-table">` utility.
- No borders between columns, only horizontal hairline dividers.
- Header text is usually `text-kumo-subtle` and capitalized.
- Hover rows change background to `bg-kumo-tint`.

### **Empty States**

- Minimalist text-only or single-graphic design.
- We avoid arbitrary icons for empty states. We rely on clear, concise
  `text-kumo-default` messages inside padded containers.

### **Modals & Dialogs**

- **Border Radius:** `rounded-xl`.
- **Backdrop:** Blurred or dark semi-transparent overlay.
- Modal headers have a subtle bottom border (`border-kumo-line`) and a standard
  `X` icon for closing.
