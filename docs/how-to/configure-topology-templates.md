# How to Configure Topology Templates

Topology templates instruct the system on how to parse your organization's
server hostnames and extract meaningful metadata, such as the Environment,
Service, and Sequence (Pod) number.

By properly configuring these templates, your dashboard will automatically
categorize incoming assets.

## The Template Editor

You can manage templates from the **Administration Panel** under the
**Topology** section. The Template Builder provides a visual interface for
constructing your parsing rules.

## Available Placeholders (Tags)

When building a template, you construct a sequence of placeholders that
represents the structure of your hostnames.

- `:env` — Extracts the **Environment** name (e.g., `prod`, `dev`, `staging`).
- `:svc` — Extracts the **Service** name (e.g., `db`, `api-gateway`,
  `frontend`).
- `:seq` — Extracts the **Sequence** or Pod number (e.g., `01`, `02`, `18`).
- `:any` — The wildcard placeholder. It tells the engine to explicitly
  **ignore** any text in this position. This is extremely useful for skipping
  data centers, region codes, or company prefixes.

> [!TIP] **Dynamic Indexing:** You can place `:env`, `:svc`, and `:seq` in **any
> arbitrary order**. The matching engine uses dynamic capture group indexing and
> will correctly assign the extracted values based on the order you define in
> the template.

## Using the Logical OR (`|`) Operator

To avoid creating dozens of identical templates for different services, you can
group multiple matching values into a single rule using the pipe character (`|`)
as a logical OR operator.

For example, if your naming convention dictates that hostnames start with the
service name followed by the environment, but you have three different APIs:

1. Navigate to the **Topology** template builder.
2. Set the template structure to: `:svc-:env-:seq`
3. In the **Service Match Value** field, input your identifiers separated by a
   pipe: `auth-api|payment-api|inventory-api`

When a hostname like `payment-api-prod-02` reports in, the engine will
dynamically evaluate the `|` delimited list,
match `payment-api` as the `:svc`, `prod` as the `:env`, and `02` as the `:seq`.

## Example Configurations

### Scenario A: Strict Environment-Service-Sequence

- **Hostname:** `prod-web-01`
- **Template Structure:** `:env-:svc-:seq`
- **Environment Value:** `prod`
- **Service Value:** `web`

### Scenario B: Ignoring Region Codes

- **Hostname:** `us-east-db-staging-05`
- **Template Structure:** `:any-:svc-:env-:seq`
- **Environment Value:** `staging`
- **Service Value:** `db`
- **Result:** The engine ignores `us-east`, identifies the service as `db`, the
  environment as `staging`, and the sequence as `05`.

### Scenario C: Grouping Multiple Environments

- **Hostname:** `cache-qa-02`
- **Template Structure:** `:svc-:env-:seq`
- **Environment Value:** `dev|qa|staging`
- **Service Value:** `cache`
- **Result:** The single template will successfully capture hostnames for the
  cache service across all three non-production environments.
