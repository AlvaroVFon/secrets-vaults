const APIKEY_BYTES = 32

// Generates a random apikey equivalent to `openssl rand -base64 32`
// (32 random bytes -> 44-character standard base64 string with padding).
export function generateApikey(bytes: number = APIKEY_BYTES): string {
  const values = new Uint8Array(bytes)
  crypto.getRandomValues(values)

  let binary = ''
  for (let i = 0; i < values.length; i++) {
    binary += String.fromCharCode(values[i])
  }

  return btoa(binary)
}
