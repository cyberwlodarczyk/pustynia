const fromBase64 = (data: string) => {
  return Uint8Array.from(atob(data), (c) => c.codePointAt(0)!)
}

const toBase64 = (data: Uint8Array) => {
  return btoa(String.fromCharCode(...data))
}

const getRandomValues = (n: number) => {
  return crypto.getRandomValues(new Uint8Array(n))
}

export const generateCode = () => {
  const bytes = getRandomValues(9)
  let code = ''
  for (let i = 0; i < 9; i++) {
    if (i === 3 || i === 6) {
      code += '-'
    }
    code += String.fromCharCode('a'.charCodeAt(0) + (bytes[i] % 26))
  }
  return code
}

export const deriveKey = async (password: string) => {
  const rawKey = await crypto.subtle.importKey(
    'raw',
    new TextEncoder().encode(password),
    { name: 'PBKDF2' },
    false,
    ['deriveKey'],
  )
  const salt = getRandomValues(32)
  const key = await crypto.subtle.deriveKey(
    { name: 'PBKDF2', hash: 'SHA-256', iterations: 10_000, salt },
    rawKey,
    { name: 'AES-GCM', length: 256 },
    false,
    ['encrypt', 'decrypt'],
  )
  return { key, salt: toBase64(salt) }
}

export const encrypt = async (key: CryptoKey, plaintext: string) => {
  const iv = getRandomValues(12)
  const buf = await crypto.subtle.encrypt(
    { name: 'AES-GCM', iv },
    key,
    new TextEncoder().encode(plaintext),
  )
  return { iv: toBase64(iv), ciphertext: toBase64(new Uint8Array(buf)) }
}

export const decrypt = async (key: CryptoKey, iv: string, ciphertext: string) => {
  const buf = await crypto.subtle.decrypt(
    { name: 'AES-GCM', iv: fromBase64(iv) },
    key,
    fromBase64(ciphertext),
  )
  return { plaintext: new TextDecoder().decode(buf) }
}
