/// <reference types="vite/client" />

// Augment env typings for our project-specific variables
interface ImportMetaEnv {
  readonly VITE_API_BASE_URL?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
