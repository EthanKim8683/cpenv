import { createClient } from "@/src/path-rpc/create-client";
import { methods } from "./methods";

export const client = createClient(methods);
