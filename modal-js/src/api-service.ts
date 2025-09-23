import type { ModalClient } from "./client";

export abstract class APIService {
  protected client: ModalClient;

  constructor(client: ModalClient) {
    this.client = client;
  }
}
