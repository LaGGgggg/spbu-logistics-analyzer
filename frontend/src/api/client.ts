import axios from 'axios'

export const apiClient = axios.create({
  baseURL: '/api',
  timeout: 10000,
  // The session cookie is HttpOnly and scoped to the API origin, so the
  // browser must attach it on every request.
  withCredentials: true,
})
