// API documentation:
//   - [resource types]
//   - [developer docs]
//
// [resource types]:
// [developer docs]: https://developers.notion.com/reference/property-object
package property

const TypeScriptSourceURL = "https://raw.githubusercontent.com/makenotion/notion-sdk-js/main/src/api-endpoints.ts"

// NOTE: I have managed to partially convert the types defined in ./api-endpoints.ts to OpenAPI format
// using `typeconv` (urn:github:grantila/typeconv).
// I encountered two issues so far:
// - the type Record is not supported (https://github.com/microsoft/TypeScript/blob/v5.0.4/src/lib/es5.d.ts#L1568)
// - references to unexported types are not included even if the option `--ts-non-exported <method>` should be `include-if-referenced` by default; I'm not sure if setting the method explicitly to `include` or `inline` will solve this problem
// - converting to SureType the `TimestampLastEditedTimeFilter` is not found even if it's present in the source
//		> npx typeconv -f ts -t st --st-missing-ref error -O - api-endpoints.ts
//		Reference to missing type "TimestampLastEditedTimeFilter"
