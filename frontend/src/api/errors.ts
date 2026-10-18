import axios from 'axios'

export function getApiErrorMessage(error: unknown): string {
  if (!axios.isAxiosError(error)) {
    return 'Произошла неизвестная ошибка'
  }

  if (error.code === 'ECONNABORTED') {
    return 'Превышено время ожидания ответа сервера'
  }

  if (!error.response) {
    return 'Не удалось подключиться к серверу'
  }

  if (error.response.status === 404) {
    return 'Запрашиваемый ресурс не найден'
  }

  if (error.response.status >= 500) {
    return 'Внутренняя ошибка сервера'
  }

  return 'Не удалось выполнить запрос'
}