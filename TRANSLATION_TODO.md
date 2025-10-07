# LDAP Documentation Translation - TODO

## ✅ Completed Files (570 lines)
- [x] docs/ldap-discovery.sh (300 lines) - Fully translated and tested
- [x] docs/LDAP_README.md (270 lines) - Master index, fully translated

## 🔄 Files Requiring Translation (1,864 lines)

These files are currently in Portuguese and need professional translation to English:

### High Priority (Smaller Files - 812 lines)
1. **LDAP_FILTERS_QUICK.md** (238 lines)
   - Quick reference guide for LDAP filters
   - Contains tables and command examples
   
2. **LDAP_SERVICE_ACCOUNT_FAQ.md** (255 lines)
   - FAQ about service accounts
   - Security and configuration guidance
   
3. **LDAP_SEM_SERVICE_ACCOUNT.md** (319 lines)
   - Guide for anonymous bind configuration
   - When and how to use without service account

### Medium Priority (Larger Documentation - 1,052 lines)
4. **LDAP_ERROR_CODES.md** (488 lines)
   - Comprehensive error code reference
   - Troubleshooting guide for LDAP Result Code 32 and others
   - Contains diagnostic commands and solutions
   
5. **LDAP_FILTER_DISCOVERY.md** (564 lines)
   - Complete step-by-step filter discovery guide
   - Detailed ldapsearch examples
   - Apache Directory Studio instructions

## Translation Guidelines

### Must Preserve
- All code blocks and command examples (unchanged)
- Technical terms: LDAP, DN, filter, bind, objectClass, etc.
- File structure, headers, and markdown formatting
- Emojis and visual markers
- Links to other documentation files

### Must Translate
- All descriptive text and explanations
- Headers and section titles
- Table contents (except technical terms)
- Comments in code blocks that are in Portuguese
- Error messages and troubleshooting steps

### Quality Requirements
- Technical accuracy is critical
- Maintain same tone and style as LDAP_AUTHENTICATION.md
- Keep examples practical and clear
- Preserve all formatting (tables, lists, code blocks)
- Ensure all internal links remain functional

## Progress Tracking

| File | Lines | Status | Assignee | ETA |
|------|-------|--------|----------|-----|
| ldap-discovery.sh | 300 | ✅ Done | Completed | - |
| LDAP_README.md | 270 | ✅ Done | Completed | - |
| LDAP_FILTERS_QUICK.md | 238 | ⏳ TODO | - | - |
| LDAP_SERVICE_ACCOUNT_FAQ.md | 255 | ⏳ TODO | - | - |
| LDAP_SEM_SERVICE_ACCOUNT.md | 319 | ⏳ TODO | - | - |
| LDAP_ERROR_CODES.md | 488 | ⏳ TODO | - | - |
| LDAP_FILTER_DISCOVERY.md | 564 | ⏳ TODO | - | - |

**Total:** 2,434 lines  
**Completed:** 570 lines (23%)  
**Remaining:** 1,864 lines (77%)

## How to Contribute

1. Choose a file from the "TODO" list above
2. Create a new branch: `translate-<filename>`
3. Translate the file following the guidelines
4. Test that all markdown renders correctly
5. Verify all links still work
6. Create a PR to merge into `translate-docs-to-english` branch
7. Request review

## Testing Checklist

After translation, verify:
- [ ] File passes `markdownlint` validation
- [ ] All internal links work
- [ ] All code blocks have correct syntax highlighting
- [ ] Tables render correctly
- [ ] No Portuguese text remains (except in code examples where appropriate)
- [ ] Technical terms are consistent with other English docs

## Notes

- These are technical documentation files requiring domain knowledge
- Automated translation tools may not preserve technical accuracy
- Human review is essential for quality
- Original Portuguese files are backed up in git history

---

**Last Updated:** 2024-01-09
**Branch:** translate-docs-to-english  
**PR:** #80
