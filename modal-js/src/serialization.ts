/**
 * CBOR serialization utilities for Modal.
 *
 * This module encapsulates cbor-x usage with a consistent configuration
 * that ensures compatibility with the Python CBOR implementation.
 */

import { Encoder, Decoder } from "cbor-x";

/**
 * Custom CBOR encoder configured for Modal's specific requirements.
 *
 * Configuration:
 * - mapsAsObjects: true - Encode Maps as Objects for compatibility
 * - useRecords: false - Disable record structures
 * - tagUint8Array: false - Don't tag Uint8Arrays (avoid tag 64)
 */
const encoder = new Encoder({
  mapsAsObjects: true,
  useRecords: false,
  tagUint8Array: false,
});

/**
 * Custom CBOR decoder configured for Modal's specific requirements.
 */
const decoder = new Decoder({
  mapsAsObjects: true,
  useRecords: false,
  tagUint8Array: false,
});

/**
 * Encode a JavaScript value to CBOR bytes.
 *
 * @param value - The JavaScript value to encode
 * @returns CBOR-encoded bytes as a Buffer
 */
export function cborEncode(value: any): Buffer {
  return encoder.encode(value);
}

/**
 * Decode CBOR bytes to a JavaScript value.
 *
 * @param data - The CBOR-encoded bytes to decode
 * @returns The decoded JavaScript value
 */
export function cborDecode(data: Buffer | Uint8Array): any {
  return decoder.decode(data);
}

/**
 * Get the configured encoder instance.
 * Useful for advanced use cases that need direct access to the encoder.
 *
 * @returns The configured CBOR encoder
 */
export function getEncoder(): Encoder {
  return encoder;
}

/**
 * Get the configured decoder instance.
 * Useful for advanced use cases that need direct access to the decoder.
 *
 * @returns The configured CBOR decoder
 */
export function getDecoder(): Decoder {
  return decoder;
}
