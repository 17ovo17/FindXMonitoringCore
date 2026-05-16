export const clearRequestAuthHeader = headers => {
  if (!headers) return
  delete headers.Authorization
  delete headers.authorization
}

const plainHeaders = headers => {
  if (!headers) return {}
  if (typeof headers.toJSON === 'function') return { ...headers.toJSON() }
  return { ...headers }
}

const dropAxiosHeaderBuckets = headers => {
  for (const key of ['common', 'get', 'post', 'put', 'patch', 'delete', 'head', 'options']) {
    delete headers[key]
  }
}

export const applyStoredAuthHeader = config => {
  const token = localStorage.getItem('aiw-token')
  const headers = plainHeaders(config.headers)
  dropAxiosHeaderBuckets(headers)
  clearRequestAuthHeader(headers)
  if (token) headers.Authorization = `Bearer ${token}`
  config.headers = headers
  return config
}

export const clearDefaultAuthHeader = axiosClient => {
  const common = axiosClient.defaults?.headers?.common
  if (!common) return
  delete common.Authorization
  delete common.authorization
}

export const setDefaultAuthHeader = (axiosClient, token) => {
  clearDefaultAuthHeader(axiosClient)
  if (!token) return
  axiosClient.defaults.headers.common.Authorization = `Bearer ${token}`
}
