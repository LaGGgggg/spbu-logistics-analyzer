import axios from 'axios'

export const apiClient = axios.create({
  baseURL: '/api',
  timeout: 10000,
})

apiClient.interceptors.request.use((config) => {
  // TODO: Добавить авторизацию после согласования с Backend.
  return config
})