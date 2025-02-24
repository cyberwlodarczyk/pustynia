<script setup>
import { onMounted } from 'vue'
import { generateCode, deriveKey, decrypt, encrypt } from '@/lib'

onMounted(async () => {
  const { key } = await deriveKey("secret")
  const { ciphertext, iv } = await encrypt(key, generateCode())
  const { plaintext } = await decrypt(key, iv, ciphertext)
  console.log(plaintext);
})
</script>

<template>
  <main>
    <h1>Hello, World!</h1>
  </main>
</template>

<style scoped>
h1 {
  font-size: 3rem;
  font-weight: 700;
  text-align: center;
}
</style>
