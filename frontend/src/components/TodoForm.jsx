import { useState } from "react";

export default function TodoForm({ isSubmitting, onSubmit }) {
  const [title, setTitle] = useState("");

  async function handleSubmit(event) {
    event.preventDefault();

    const trimmedTitle = title.trim();
    if (!trimmedTitle) {
      return;
    }

    await onSubmit(trimmedTitle);
    setTitle("");
  }

  return (
    <form className="todo-form" onSubmit={handleSubmit}>
      <input
        aria-label="New todo title"
        placeholder="Add a new task"
        type="text"
        value={title}
        onChange={(event) => setTitle(event.target.value)}
      />
      <button className="button" disabled={isSubmitting || !title.trim()} type="submit">
        Add
      </button>
    </form>
  );
}
