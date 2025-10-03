-- Postgres schema for yearly custcode snapshot and monthly details.

-- Stores top-200 custcodes per branch for a fiscal year (start in October).
CREATE TABLE IF NOT EXISTS bm_custcode_init (
    id BIGSERIAL PRIMARY KEY,
    fiscal_year INT NOT NULL,
    branch_code TEXT NOT NULL,
    org_name TEXT,
    cust_code TEXT NOT NULL,
    use_type TEXT,
    use_name TEXT,
    cust_name TEXT,
    address TEXT,
    route_code TEXT,
    meter_no TEXT,
    meter_size TEXT,
    meter_brand TEXT,
    meter_state TEXT,
    debt_ym TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (fiscal_year, branch_code, cust_code)
);

-- Stores monthly details for custcodes captured above.
CREATE TABLE IF NOT EXISTS bm_meter_details (
    id BIGSERIAL PRIMARY KEY,
    year_month TEXT NOT NULL,
    branch_code TEXT NOT NULL,
    org_name TEXT,
    cust_code TEXT NOT NULL,
    use_type TEXT,
    use_name TEXT,
    cust_name TEXT,
    address TEXT,
    route_code TEXT,
    meter_no TEXT,
    meter_size TEXT,
    meter_brand TEXT,
    meter_state TEXT,
    average NUMERIC,
    present_meter_count NUMERIC,
    present_water_usg NUMERIC,
    debt_ym TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (year_month, branch_code, cust_code)
);
