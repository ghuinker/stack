import axios from 'axios'
import { urlToApiUrl } from './api'

const createAxios = () => {
  let http = axios.create({
    baseURL: '/api'
  })

  http.defaults.xsrfHeaderName = 'X-CSRFToken'
  http.defaults.xsrfCookieName = 'csrftoken'
  http.defaults.withCredentials = true

  http.defaults.headers.common['Accept'] = 'application/json'
  http.defaults.headers.common['Cache-Control'] = 'no-cache'
  http.defaults.headers.common['Pragma'] = 'no-cache' // IOS
  http.defaults.headers.common['Expires'] = '0'

  // http.defaults.notify = true
  // @ts-ignore
  http.interceptors.request.use(config => {
    try {
      config.url = urlToApiUrl(config.url ?? '')
    } catch (error) {
      console.error(error)
      return null
    }
    return config
  })
  return http
}

const UNAUTHORIZED = 401
const NOT_ALLOWED = 403
const NOT_FOUND = 404
const INVALID_DATA = 422

const NOTIFY_SUCCESS_METHODS = ['post', 'patch', 'delete']

export const mittError = (mitt, error) => {
  const { status, data } = error.response

  const detail = data.detail
  if (status == NOT_ALLOWED || status == UNAUTHORIZED) {
    mitt.emit('notify', {
      level: 'Warning',
      title: detail ?? 'Action not allowed',
      description: data.description ?? null
    })
  } else if (status == NOT_FOUND) {
    mitt.emit('notify', {
      level: 'Error',
      title: detail ?? 'Resource not Found'
    })
  } else if (status == INVALID_DATA) {
    mitt.emit('notify', {
      level: 'Error',
      title: detail ?? 'Invalid data',
      description: data.description ?? null
    })
  } else {
    mitt.emit('notify', {
      level: 'Error',
      title: detail ?? 'There was an error',
      description: data.description ?? null
    })
  }
  // Prevent unwanted rendering of error list
  return Promise.reject(data)
}

export const useAxios = ({ onResponse = (r: { data: any }) => r.data }) => {
  const http = createAxios()
  http.interceptors.response.use(onResponse, error => console.error(error))
  return http
}

export const useMittAxios = (mitt, { onResponse = (r: { data: any }) => r.data }) => {
  const http = createAxios()
  http.interceptors.response.use(onResponse, error => mittError(mitt, error))
  return http
}

export default useAxios
