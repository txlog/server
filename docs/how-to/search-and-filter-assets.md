# How to Search and Filter Assets

This guide explains how to use the search and filtering capabilities in the Txlog Server dashboard to quickly locate
specific assets and identify stale or inactive systems.

## 🎯 Overview

The Assets page provides a central search bar and quick filters to help you manage large fleets of servers. You can
search by hostname, machine ID, or use special keywords to find systems based on their status.

## 🛠️ How-to: Finding Specific Assets

### Basic Search

1. Navigate to the **Assets** page from the main dashboard.
2. In the search bar at the top, type the **hostname** or **Machine ID** of the asset you're looking for.
3. Press **Enter** or click the search icon.

### Using Quick Filters

You can use the toggle icons next to the search bar for immediate filtering:

- **Needs Restart** (Golden Icon): Shows only assets that require a reboot after updates.
- **Inactive** (Coral Icon): Shows only assets that haven't synchronized data in the last 15 days.

---

## 📖 Reference: Search Keywords

The search bar supports "magic keywords" that allow for advanced filtering directly from the text input. These can be
typed manually in the search field.

| Keyword         | Description                                                           |
| :-------------- | :-------------------------------------------------------------------- |
| `restart:true`  | Filters assets that require a system restart.                         |
| `inactive:true` | Filters assets that have been inactive for more than 15 days.         |

> [!TIP] You can combine text with keywords. For example, searching `prod-server restart:true` will find production
> servers that require a restart.

---

## 💡 Explanation: How Search Works

### Matching Logic

When you enter a search term:

1. **Keyword Interception**: The server first checks if the query contains any special keywords (like `restart:true`).
   If found, it applies the specific database filter.
2. **Text Matching**: If no keywords are present (or for the remaining text), the server performs a case-insensitive
   partial match against the **hostname** and an exact match against the **Machine ID**.

### Visual Status Indicators

Assets in the list display colored badges on their OS icon to indicate their state:

- **Golden Pulse**: System restart required.
