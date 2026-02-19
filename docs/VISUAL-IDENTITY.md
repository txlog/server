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
for categorization and highlights.

### **Primary Colors**

These form the foundation of the Txlog UI.

* **Indigo Slate (Primary Dark):** \#424565
  * *Usage:* Text, headers, main navigation, sidebar backgrounds, and
    high-emphasis borders (inspired by the notebook spine and pencil lead).
* **Lavender Frost (Primary Light):** \#E6E6FA (Approximate)
  * *Usage:* App backgrounds, large card surfaces, and subtle interactive states
    (inspired by the notebook cover).
* **White (Base Surface):** \#FFFFFF
  * *Usage:* Primary container backgrounds to ensure high contrast and
    readability.

### **Accent & Semantic Colors**

Used for actions, status indicators, and categorization tags.

* **Coral Bookmark (Action/Destructive):** \#D9556A
  * *Usage:* Primary call-to-action (CTA) buttons, error states, and important
    highlights.
* **Golden Label (Warning/Highlight):** \#F4B54B
  * *Usage:* Warning alerts, badges, highlighted text, and secondary CTAs.
* **Sky Blue (Info/Active):** \#6AA2FB
  * *Usage:* Active states, text links, informational banners, and progress
    bars.
* **Leaf Green (Success):** \#6DB865
  * *Usage:* Success messages, "completed" statuses, and positive trends.
* **Peach Wood (Tertiary/Subtle):** \#F1A994
  * *Usage:* Subtle background highlights, illustrations, and empty states.

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
  (\#6AA2FB).
  * Text Color: White.
  * Border Radius: 12px.
  * Hover State: 10% darker shade with a slight upward lift transform
    (translateY(-1px)).
* **Secondary Button:**
  * Background: Transparent.
  * Border: 2px solid Indigo Slate (\#424565).
  * Text Color: Indigo Slate (\#424565).
  * Border Radius: 12px.

### **Input Fields**

* Background: White or very light gray.
* Border: 2px solid Lavender Frost (\#E6E6FA).
* Focus State: Border changes to Sky Blue (\#6AA2FB) with a soft glow
  (box-shadow).
* Border Radius: 12px.

### **Cards & Containers**

* Background: White.
* Border Radius: 16px or 24px.
* Shadow: Soft, diffused drop shadow to create depth without making it look
  heavy.
  * *CSS Example:* box-shadow: 0 8px 24px rgba(66, 69, 101, 0.08);

### **Tabs & Navigation (Inspired by the Notebook Tabs)**

* Use the Accent colors (Blue, Green) for navigation tabs.
* Active tab should pop out slightly, mimicking a physical bookmark or index tab
  protruding from a book.
* Inactive tabs should be muted or grayed out.
