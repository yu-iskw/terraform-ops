# Terraform Plan Summary

**Plan Status:** ❌ Not Applicable  
**Format Version:** 1.2  
**Complete:** true

## 📊 Statistics

**Total Changes:** 8

### By Action

- ❌ **delete:** 6
- ➕ **create:** 7

### By Provider

- 🏢 **random:** 8

### By Module

- 📦 **Root Module:** 6
- 📦 **module.myrandom:** 2

## 🔄 Resource Changes

### ➕ Create (2)

- **module.myrandom.module.myrandom.random_integer.test_integer**
- **module.myrandom.module.myrandom.random_string.test_string**

### 🔄 Replace (5)

- **random_id.test_id**
- **random_integer.test_integer**
- **random_password.test_password**
  - 🔒 Contains sensitive values
- **random_string.test_string**
- **random_uuid.test_uuid**

### ❌ Delete (1)

- **random_pet.test_pet**

## 📤 Output Changes

- **test_integer**
- **test_prefix**
- **test_uuid**
- **test_id**
- **test_password**
- **test_pet**
- **test_string**
- **test_tag**
- **myrandom**
