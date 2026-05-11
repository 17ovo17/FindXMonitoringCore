import { get, post, put, del } from './http.js'

export const cmdbApi = {
  tree: () => get('/cmdb/tree'),
  objects: {
    list: (categoryId) => get('/cmdb/objects', { params: categoryId ? { category_id: categoryId } : {} }),
    get: (id) => get(`/cmdb/objects/${id}`),
    create: (body) => post('/cmdb/objects', body),
    update: (id, body) => put(`/cmdb/objects/${id}`, body),
    remove: (id) => del(`/cmdb/objects/${id}`),
  },
  attributes: {
    list: (objectId) => get(`/cmdb/objects/${objectId}/attributes`),
    create: (objectId, body) => post(`/cmdb/objects/${objectId}/attributes`, body),
    update: (id, body) => put(`/cmdb/attributes/${id}`, body),
    remove: (id) => del(`/cmdb/attributes/${id}`),
  },
  instances: {
    list: (objectId, params) => get(`/cmdb/objects/${objectId}/instances`, { params }),
    get: (id) => get(`/cmdb/instances/${id}`),
    create: (objectId, body) => post(`/cmdb/objects/${objectId}/instances`, body),
    update: (id, body) => put(`/cmdb/instances/${id}`, body),
    remove: (id) => del(`/cmdb/instances/${id}`),
  },
}
