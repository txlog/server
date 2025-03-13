CREATE TABLE "transactions" (
  "transaction_id" INTEGER,
  "machine_id" TEXT,
  "hostname" TEXT,
  "begin_time" TIMESTAMP WITH TIME ZONE,
  "end_time" TIMESTAMP WITH TIME ZONE,
  "actions" TEXT,
  "altered" TEXT,
  "user" TEXT,
  "return_code" TEXT,
  "release_version" TEXT,
  "command_line" TEXT,
  "comment" TEXT,
  "scriptlet_output" TEXT,
  PRIMARY KEY ("transaction_id", "machine_id")
);

CREATE TABLE "transaction_items" (
  "item_id" SERIAL PRIMARY KEY,
  "transaction_id" INTEGER,
  "machine_id" TEXT,
  "action" TEXT,
  "package" TEXT,
  "version" TEXT,
  "release" TEXT,
  "epoch" TEXT,
  "arch" TEXT,
  "repo" TEXT,
  "from_repo" TEXT
);

ALTER TABLE "transaction_items" ADD FOREIGN KEY ("transaction_id", "machine_id") REFERENCES "transactions" ("transaction_id", "machine_id");

CREATE TABLE "executions" (
  "id" SERIAL PRIMARY KEY,
  "machine_id" text NOT NULL,
  "hostname" text NOT NULL,
  "executed_at" timestamp with time zone NOT NULL,
  "success" boolean NOT NULL,
  "details" text,
  "transactions_processed" integer,
  "transactions_sent" integer
);
