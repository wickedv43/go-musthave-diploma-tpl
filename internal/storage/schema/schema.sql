CREATE TABLE "users" (
  "id" serial PRIMARY KEY,
  "login" VARCHAR(255) UNIQUE NOT NULL,
  "password" VARCHAR(255) NOT NULL,
  "balance_current" INTEGER NOT NULL DEFAULT 500,
  "balance_withdrawn" INTEGER NOT NULL DEFAULT 500
);

CREATE TABLE "orders" (
  "number" VARCHAR(255) PRIMARY KEY,
  "user_id" INTEGER NOT NULL,
  "status" VARCHAR(50) NOT NULL,
  "accrual" INTEGER NOT NULL,
  "uploaded_at" TIMESTAMPTZ NOT NULL DEFAULT (now())
);

CREATE TABLE "bills" (
  "id" SERIAL PRIMARY KEY,
  "order_number" VARCHAR(255) NOT NULL,
  "user_id" INTEGER NOT NULL,
  "sum" INTEGER NOT NULL,
  "processed_at" TIMESTAMPTZ NOT NULL DEFAULT (now())
);

ALTER TABLE "orders" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "bills" ADD FOREIGN KEY ("order_number") REFERENCES "orders" ("number");

ALTER TABLE "bills" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");
