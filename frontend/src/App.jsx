import { useEffect, useState } from "react";

import AuthForm from "./components/AuthForm.jsx";
import TodoForm from "./components/TodoForm.jsx";
import TodoList from "./components/TodoList.jsx";
import {
  clearStoredToken,
  createTodo,
  deleteTodo,
  getCurrentUser,
  getStoredToken,
  getTodos,
  login,
  register,
  setStoredToken,
  updateTodo
} from "./api.js";

export default function App() {
  const [token, setToken] = useState(() => getStoredToken());
  const [user, setUser] = useState(null);
  const [todos, setTodos] = useState([]);
  const [isCheckingSession, setIsCheckingSession] = useState(Boolean(token));
  const [isSubmittingAuth, setIsSubmittingAuth] = useState(false);
  const [isLoadingTodos, setIsLoadingTodos] = useState(false);
  const [isCreatingTodo, setIsCreatingTodo] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    let ignore = false;

    async function restoreSession() {
      if (!token) {
        setIsCheckingSession(false);
        return;
      }

      try {
        const [sessionData, todosData] = await Promise.all([
          getCurrentUser(token),
          getTodos()
        ]);

        if (!ignore) {
          setUser(sessionData.user);
          setTodos(todosData.todos);
        }
      } catch {
        if (!ignore) {
          clearStoredToken();
          setToken(null);
          setUser(null);
          setTodos([]);
        }
      } finally {
        if (!ignore) {
          setIsCheckingSession(false);
        }
      }
    }

    restoreSession();

    return () => {
      ignore = true;
    };
  }, [token]);

  async function handleAuthSubmit(mode, username, password) {
    setError("");
    setIsSubmittingAuth(true);

    try {
      const data = mode === "register"
        ? await register(username, password)
        : await login(username, password);

      setStoredToken(data.token);
      setToken(data.token);
      setUser(data.user);
      await loadTodos();
    } catch (err) {
      setError(err.message);
    } finally {
      setIsSubmittingAuth(false);
    }
  }

  function handleLogout() {
    clearStoredToken();
    setToken(null);
    setUser(null);
    setTodos([]);
    setError("");
  }

  async function loadTodos() {
    setIsLoadingTodos(true);
    setError("");

    try {
      const data = await getTodos();
      setTodos(data.todos);
    } catch (err) {
      setError(err.message);
    } finally {
      setIsLoadingTodos(false);
    }
  }

  async function handleCreateTodo(title) {
    setIsCreatingTodo(true);
    setError("");

    try {
      const data = await createTodo(title);
      setTodos((currentTodos) => [data.todo, ...currentTodos]);
    } catch (err) {
      setError(err.message);
    } finally {
      setIsCreatingTodo(false);
    }
  }

  async function handleToggleTodo(todo) {
    setError("");

    try {
      const data = await updateTodo(todo.id, todo.title, !todo.completed);
      replaceTodo(data.todo);
    } catch (err) {
      setError(err.message);
    }
  }

  async function handleUpdateTodoTitle(todo, title) {
    setError("");

    try {
      const data = await updateTodo(todo.id, title, todo.completed);
      replaceTodo(data.todo);
    } catch (err) {
      setError(err.message);
    }
  }

  async function handleDeleteTodo(todo) {
    setError("");

    try {
      await deleteTodo(todo.id);
      setTodos((currentTodos) => currentTodos.filter((item) => item.id !== todo.id));
    } catch (err) {
      setError(err.message);
    }
  }

  function replaceTodo(updatedTodo) {
    setTodos((currentTodos) =>
      currentTodos.map((todo) => (todo.id === updatedTodo.id ? updatedTodo : todo))
    );
  }

  if (isCheckingSession) {
    return (
      <main className="app-shell app-shell--centered">
        <section className="panel auth-panel">
          <p className="status-text">Checking session...</p>
        </section>
      </main>
    );
  }

  if (!user) {
    return (
      <main className="app-shell app-shell--centered">
        <AuthForm
          error={error}
          isSubmitting={isSubmittingAuth}
          onSubmit={handleAuthSubmit}
        />
      </main>
    );
  }

  return (
    <main className="app-shell">
      <header className="topbar">
        <div>
          <p className="eyebrow">Signed in as</p>
          <h1>{user.username}</h1>
        </div>
        <button className="button button--secondary" type="button" onClick={handleLogout}>
          Logout
        </button>
      </header>

      <section className="panel">
        <div className="dashboard-header">
          <div>
            <p className="eyebrow">Tasks</p>
            <h2>Today&apos;s work</h2>
          </div>
          <p className="todo-count">
            {todos.filter((todo) => !todo.completed).length} active
          </p>
        </div>

        <TodoForm isSubmitting={isCreatingTodo} onSubmit={handleCreateTodo} />

        {error ? <p className="error-text dashboard-error">{error}</p> : null}

        {isLoadingTodos ? (
          <p className="status-text">Loading tasks...</p>
        ) : (
          <TodoList
            todos={todos}
            onDelete={handleDeleteTodo}
            onToggle={handleToggleTodo}
            onUpdateTitle={handleUpdateTodoTitle}
          />
        )}
      </section>
    </main>
  );
}
