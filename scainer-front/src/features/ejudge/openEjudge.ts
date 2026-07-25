import type { EjudgeBrowserLogin } from "@/client/types.gen";

/** POST-логин в ejudge в браузере → страница контеста. */
export function openEjudgeContest(login: EjudgeBrowserLogin): void {
  const win = window.open("about:blank", "_blank");
  if (!win) {
    return;
  }
  const doc = win.document;
  const form = doc.createElement("form");
  form.method = "POST";
  form.action = `${login.base_url.replace(/\/$/, "")}/cgi-bin/new-master`;
  form.target = "_self";
  const fields: Record<string, string> = {
    login: login.login,
    password: login.password,
    contest_id: String(login.contest_id),
    role: "6",
    action_2: "Submit",
  };
  for (const [name, value] of Object.entries(fields)) {
    const input = doc.createElement("input");
    input.type = "hidden";
    input.name = name;
    input.value = value;
    form.appendChild(input);
  }
  doc.body.appendChild(form);
  form.submit();
}
