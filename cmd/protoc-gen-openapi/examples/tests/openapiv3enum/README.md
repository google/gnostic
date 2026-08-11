# OpenAPI Enumeration Feature Improvements

## Overview

We have improved the enumeration support in the `protoc-gen-openapi` tool to correctly generate OpenAPI documentation containing enumeration values, with support for handling naming conflicts in nested enumerations.

## Major Improvements

1. **Unified Enumeration Value Generation**: Enumeration values are generated regardless of what `enum_type` is set to
2. **Field Name Usage**: Enumeration values use the field names defined in proto, rather than numeric values
3. **Type Unification**: Changed the default type from "integer" to "string" since we're using enumeration field names
4. **Backward Compatibility**: Still supports the `enum_type` parameter to control behavior
5. **Nested Enumeration Support**: Correctly handles enumerations nested within messages, avoiding naming conflicts
6. **Naming Conflict Resolution**: Nested enumerations use the `ParentMessage_EnumName` format to avoid conflicts

## Nested Enumeration Handling

### Naming Rules

- **File-level enumerations**: `EnumName` (e.g., `UserStatus`)
- **Nested enumerations**: `ParentMessage_EnumName` (e.g., `User_Status`)

### Example

```protobuf
message User {
  enum Status {  // Nested enumeration
    UNKNOWN = 0;
    ACTIVE = 1;
    INACTIVE = 2;
  }
  Status status = 1;
}

enum UserRole {  // File-level enumeration
  ADMIN = 0;
  USER = 1;
}
```

Generated OpenAPI Schema:

```yaml
components:
  schemas:
    User_Status:  # Nested enumeration, using concatenated name
      type: string
      format: enum
      enum:
        - UNKNOWN
        - ACTIVE
        - INACTIVE
    
    UserRole:     # File-level enumeration, using original name
      type: string
      format: enum
      enum:
        - ADMIN
        - USER
    
    User:
      type: object
      properties:
        status:
          $ref: '#/components/schemas/User_Status'  # Reference to nested enumeration
```

## Important Notes

1. Enumeration values use the field names defined in proto, not numeric values
2. This ensures the generated OpenAPI documentation is clearer and more readable
3. Maintains compatibility with existing code
4. Nested enumerations automatically handle naming conflicts using the `ParentMessage_EnumName` format
5. All enumeration types are defined in `components/schemas`, with fields referencing them via `$ref`
