import { useState } from "react";

export default function AuthForm({ error, isSubmitting, onSubmit }) {
  const [mode, setMode] = useState("login");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");

  const isRegister = mode === "register";

  function handleSubmit(event) {
    event.preventDefault();
    onSubmit(mode, username.trim(), password);
  }

  return (
    <section className="panel auth-panel">
      <div className="auth-header">
        <div>
          <p className="eyebrow">ToDo App</p>
          <h1>{isRegister ? "Create account" : "Welcome back"}</h1>
        </div>

        <div className="segmented-control" aria-label="Authentication mode">
          <button
            className={mode === "login" ? "active" : ""}
            type="button"
            onClick={() => setMode("login")}
          >
            Login
          </button>
          <button
            className={mode === "register" ? "active" : ""}
            type="button"
            onClick={() => setMode("register")}
          >
            Register
          </button>
        </div>
      </div>

      <form className="form" onSubmit={handleSubmit}>
        <label>
          Username
          <input
            autoComplete="username"
            minLength={3}
            required
            type="text"
            value={username}
            onChange={(event) => setUsername(event.target.value)}
          />
        </label>

        <label>
          Password
          <input
            autoComplete={isRegister ? "new-password" : "current-password"}
            minLength={8}
            required
            type="password"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
          />
        </label>

        {error ? <p className="error-text">{error}</p> : null}

        <button className="button" disabled={isSubmitting} type="submit">
          {isSubmitting ? "Please wait..." : isRegister ? "Create account" : "Login"}
        </button>
      </form>
    </section>
  );
}
