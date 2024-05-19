import { ref, onMounted, watch, Ref } from 'vue'
import useAxios from './axios.js'
import { AxiosResponse } from 'axios'

interface ApiOptions {
  params?: Record<string, any> | (() => Record<string, any>)
  enabled?: boolean
  initialData?: any
  requestOnMount?: boolean
  onSuccess?: (data: any) => void
}

interface ApiResult<T> {
  isLoading: Ref<boolean>
  data: Ref<T | null>
  request: () => Promise<void>
  response: Ref<AxiosResponse<T> | null>
}

export function urlToApiUrl(url: string, addApiPrepend = false) {
  const pathname = window.location.pathname
  let returnUrl = url
  // If no url given, default to current path
  if (!url) {
    returnUrl = pathname
  } else if (url.startsWith('//')) {
    returnUrl = url.substring(1, url.length)
    if (addApiPrepend) returnUrl = '/api' + returnUrl
    returnUrl += returnUrl.endsWith('/') ? '' : '/'
    return returnUrl
  }
  // Does not start with slash -> append to current path
  else if (!url.startsWith('/')) {
    const connectingSlash = pathname.endsWith('/') ? '' : '/'
    returnUrl = pathname + connectingSlash + url
  }

  if (addApiPrepend) returnUrl = '/api' + returnUrl
  return returnUrl
}

function _getApi<T>(
  url: string,
  { params = {}, enabled = true, initialData = null, requestOnMount = true, onSuccess = (data: T) => {} }
): ApiResult<T> {
  const data: Ref<T | null> = ref(initialData)
  const isLoading: Ref<boolean> = ref(false)
  const response: Ref<AxiosResponse<T, any> | null> = ref(null)
  const axios = useAxios({ onResponse: (r: any) => r })

  const request = async () => {
    if (!enabled) return
    const internalParams = params instanceof Function ? params() : params
    isLoading.value = true
    axios
      .get(url, {
        params: internalParams
      })
      .then((res: AxiosResponse<T>) => {
        isLoading.value = false
        response.value = res
        data.value = res.data
        onSuccess(res.data)
      })
  }

  watch(() => params, request)
  watch(() => url, request)

  onMounted(() => {
    if (requestOnMount) request()
  })

  return {
    isLoading,
    data,
    request,
    response
  }
}

// TODO: fix this up
interface ApiOptions {
  params?: Record<string, any> | (() => Record<string, any>)
  enabled?: boolean
  initialData?: any
  requestOnMount?: boolean
  onSuccess?: (data: any) => void
}

export function getApi<T>(url: string, options: ApiOptions | undefined = undefined): ApiResult<T> {
  const defaultOptions: ApiOptions = {
    params: {},
    enabled: true,
    initialData: null,
    requestOnMount: true,
    onSuccess: (data: T) => {}
  }

  const mergedOptions: ApiOptions = {
    ...defaultOptions,
    ...(options ?? {})
  }
  return _getApi<T>(url, mergedOptions)
}

export default {
  get: getApi
}
