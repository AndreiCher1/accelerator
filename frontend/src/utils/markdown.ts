/**
 * Минималистичный markdown → HTML рендер для конспектов.
 * Поддерживает заголовки `#..######`, упорядоченные/неупорядоченные списки
 * (с вложенностью по отступам) и жирный/курсивный inline-текст.
 *
 * Достаточно для текстов, которые отдаёт воркер суммаризации; не претендует
 * на CommonMark-совместимость. Источник — наш бекенд, поэтому экранируем HTML
 * перед инлайн-форматированием, но v-html здесь безопасен.
 */

const ListType = {
  Unordered: "ul",
  Ordered: "ol",
} as const;

type ListType = (typeof ListType)[keyof typeof ListType];

interface ListFrame {
  type: ListType;
  indent: number;
}

function escapeHtml(value: string): string {
  return value
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;");
}

function formatInline(value: string): string {
  return escapeHtml(value)
    .replace(/\*\*(.+?)\*\*/g, "<strong>$1</strong>")
    .replace(/(^|[^*])\*(?!\s)([^*]+?)\*(?!\*)/g, "$1<em>$2</em>")
    .replace(/`([^`]+?)`/g, "<code>$1</code>");
}

export function renderMarkdown(source: string | null | undefined): string {
  if (!source) return "";

  const lines = source.split(/\r?\n/);
  const out: string[] = [];
  const listStack: ListFrame[] = [];
  let inParagraph = false;

  const closeListsTo = (indent: number) => {
    while (listStack.length > 0 && listStack[listStack.length - 1].indent >= indent) {
      const top = listStack.pop()!;
      out.push(top.type === ListType.Unordered ? "</ul>" : "</ol>");
    }
  };

  const closeAllLists = () => {
    while (listStack.length > 0) {
      const top = listStack.pop()!;
      out.push(top.type === ListType.Unordered ? "</ul>" : "</ol>");
    }
  };

  const closeParagraph = () => {
    if (inParagraph) {
      out.push("</p>");
      inParagraph = false;
    }
  };

  for (const rawLine of lines) {
    const line = rawLine.replace(/\s+$/u, "");

    if (line.trim() === "") {
      closeParagraph();
      closeAllLists();
      continue;
    }

    const heading = /^(#{1,6})\s+(.+)$/.exec(line);
    if (heading) {
      closeParagraph();
      closeAllLists();
      const level = heading[1].length;
      out.push(`<h${level}>${formatInline(heading[2])}</h${level}>`);
      continue;
    }

    const unordered = /^(\s*)[-*+]\s+(.+)$/.exec(line);
    const ordered = /^(\s*)(\d+)\.\s+(.+)$/.exec(line);

    if (unordered || ordered) {
      closeParagraph();
      const indent = (unordered ? unordered[1] : ordered![1]).length;
      const content = unordered ? unordered[2] : ordered![3];
      const desired = unordered ? ListType.Unordered : ListType.Ordered;

      while (
        listStack.length > 0 &&
        listStack[listStack.length - 1].indent > indent
      ) {
        const top = listStack.pop()!;
        out.push(top.type === ListType.Unordered ? "</ul>" : "</ol>");
      }

      const top = listStack[listStack.length - 1];
      if (!top || top.indent < indent) {
        listStack.push({ type: desired, indent });
        out.push(desired === ListType.Unordered ? "<ul>" : "<ol>");
      } else if (top.type !== desired) {
        // Тот же отступ, но другой тип — переоткрываем список.
        out.push(top.type === ListType.Unordered ? "</ul>" : "</ol>");
        listStack[listStack.length - 1] = { type: desired, indent };
        out.push(desired === ListType.Unordered ? "<ul>" : "<ol>");
      }

      out.push(`<li>${formatInline(content)}</li>`);
      continue;
    }

    // Обычный абзац: накапливаем строки.
    closeListsTo(0);
    if (!inParagraph) {
      out.push("<p>");
      inParagraph = true;
    } else {
      out.push(" ");
    }
    out.push(formatInline(line.trim()));
  }

  closeParagraph();
  closeAllLists();

  return out.join("");
}
