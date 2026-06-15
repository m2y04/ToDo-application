const API_URL = import.meta.env.VITE_API_URL || "http://localhost:5000";

export function getStoredToken() {
  return localStorage.getItem("todo_token");
}

export function setStoredToken(token) {
  localStorage.setItem("todo_token", token);
}

export function clearStoredToken() {
  localStorage.removeItem("todo_token");
}

export async function request(path, options = {}) {
  const token = options.token ?? getStoredToken();
  const headers = {
    "Content-Type": "application/json",
    ...(options.headers || {})
  };

  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  const response = await fetch(`${API_URL}${path}`, {
    ...options,
    headers
  });

  const text = await response.text();
  const data = text ? JSON.parse(text) : null;

  if (!response.ok) {
    throw new Error(data?.error || "Request failed");
  }

  return data;
}

export function register(username, password) {
  return request("/auth/register", {
    method: "POST",
    body: JSON.stringify({ username, password }),
    token: ""
  });
}

export function login(username, password) {
  return request("/auth/login", {
    method: "POST",
    body: JSON.stringify({ username, password }),
    token: ""
  });
}

export function getCurrentUser(token) {
  return request("/auth/me", { token });
}

export function getTodos() {
  return request("/todos");
}

export function createTodo(title) {
  return request("/todos", {
    method: "POST",
    body: JSON.stringify({ title })
  });
}

export function updateTodo(id, title, completed) {
  return request(`/todos/${id}`, {
    method: "PUT",
    body: JSON.stringify({ title, completed })
  });
}

export function deleteTodo(id) {
  return request(`/todos/${id}`, {
    method: "DELETE"
  });
}
