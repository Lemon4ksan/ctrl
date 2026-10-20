<img src="assets/logo.svg" align="right" width="128" height="128" alt="ctrl logo" />

# ctrl

_«В терминале задержка и визуальный шум — это состояние по умолчанию. Пусть ctrl станет вашей точной командной поверхностью без единой лишней зависимости»_

[![Go Version](https://img.shields.io/badge/go-1.27%2B-007d9c?logo=go&logoColor=white&style=flat-square)](https://go.dev/)
[![Go Reference](https://img.shields.io/badge/godoc-reference-007d9c?style=flat-square)](https://pkg.go.dev/github.com/lemon4ksan/ctrl)
[![Architecture](https://img.shields.io/badge/arch-Zero--Alloc%20%7C%20POSIX%20%7C%20MCP-0d9488?style=flat-square)](#правила)
[![Ecosystem](https://img.shields.io/badge/ecosystem-foundation-0284c7?style=flat-square)](https://github.com/lemon4ksan/foundation)
[![License](https://img.shields.io/badge/license-BSD--3--Clause-blue?style=flat-square)](LICENSE)

**[English](README.md)** | **[Русский](README_RU.md)** | **[Концепция архитектуры](docs/CONCEPT.md)**

Фреймворк командной строки, терминального интерфейса (`tuikit`), маршрутизатор Model Context Protocol v1.0 (`mcpkit`) и движок Windows UI Automation (`uiakit`) для Go 1.27+. Построен поверх стандартной библиотеки Go и [`foundation`](https://github.com/lemon4ksan/foundation) без сторонних зависимостей.

## Инженерные компромиссы

| Стандартная библиотека / `cobra` / `lipgloss` | Что делает `ctrl` | Цена решения |
| :--- | :--- | :--- |
| `flag.FlagSet` останавливается на первом позиционном аргументе, не поддерживает склеивание `-la` и не подсказывает опечатки | [`argkit`](argkit) нормализует токены оболочки и разбирает короткие/длинные POSIX-флаги на стеке перед передачей в `flag.FlagSet`, предлагая исправления по Левенштейну (`0 allocs/op`) | Стековые буферы расстояния Левенштейна ограничены именами флагов длиной до 64 рун |
| Полноэкранные TUI-фреймворки навязывают цикл событий Elm и захват экрана даже для простых CLI-утилит | [`tuikit`](tuikit) потоково рендерит таблицы, карточки, прогресс-бары, спарклайны и бейджи напрямую в `io.Writer`, автоматически переключаясь на чистый лог при `NO_COLOR` или перенаправлении в пайп | Не предназначен для полноэкранных диалоговых окон с альтернативным буфером терминала |
| Ручная реализация MCP-сервера требует вложенных `map[string]any` для JSON Schema и ручного разбора JSON-RPC | [`mcpkit`](mcpkit) предоставляет декларативный конструктор `NewTool`, обобщённые адаптеры `TypedHandler[T]` и транспорты STDIO / HTTP+SSE с буфером 16 МБ и Token Bucket лимитером | Ориентирован строго на спецификацию MCP `2024-11-05` (`v1.0`), а не на произвольные диалекты JSON-RPC |
| Инспекция дерева доступности в Windows обычно требует обёрток на C++/C# или CGO | [`uiakit`](uiakit) вызывает виртуальные таблицы COM `CUIAutomation8` (`IUIAutomation2`) напрямую через `foundation/silicon/comkit` с пакетной предвыборкой `IUIAutomationCacheRequest` | Специфичен для Windows (`UIAutomationCore.dll`); на других ОС компилируется в предсказуемую заглушку |

## Правила

1. **Ноль сторонних зависимостей.** `ctrl` опирается только на стандартную библиотеку Go и [`foundation`](https://github.com/lemon4ksan/foundation) — без рефлексивных мапперов тегов и транзитивных деревьев зависимостей.
2. **Ноль аллокаций на горячем пути.** Подсчёт визуальной ширины моноширинных строк (`VisibleWidth`), очистка ANSI-кодов, push-итератор `LinesSeq` и подсказки опечаток Левенштейна работают с `0 B/op`.
3. **Явный ввод-вывод и изолированная стилизация.** Обработчики команд явно принимают `(ctx, args, stdout, stderr)`, а экземпляры `tuikit.Styler` хранят политику цвета локально без мутации глобального состояния.
4. **Автоматическая деградация в пайпах и CI.** Прогресс-бары и ANSI-последовательности проверяют интерактивность терминала (`ProbeTerminal`) и в неинтерактивных потоках переходят на чистый построчный вывод.

## Структура

- [`ctrl` (корень)](ctrl.go) — Корневой фасад, реэкспортирующий `App`, `Command`, `GroupCommand`, `HubCommand`, `ParseInterspersedFlags`, `Table`, `Box`, `Progress` и статусные бейджи.
- [`argkit`](argkit) — Движок POSIX-аргументов: свободное перемежение флагов и позиционных параметров (`app -v src/ -o=bin dst/`), склеивание коротких флагов (`-la` $\to$ `-l -a`), прикреплённые значения (`-I*.tmp` $\to$ `-I=*.tmp`), терминатор `--`, нормализация кавычек PowerShell/CMD (`NormalizeArgs`) и нечёткий поиск опечаток.
- [`tuikit`](tuikit) — Маршрутизатор подкоманд (`App`, `CommandGroup`), детектор возможностей терминала (`ProbeTerminal`), автовыравниваемые таблицы `Table`, рамочные карточки `Box` (`BorderSingle`, `BorderRounded`, `BorderHeavy`), живой индикатор `Progress` со скоростью и ETA, спарклайны (`RenderSparkline`: ` ▂▃▅▆▇█`) и бейджи пайплайнов.
- [`mcpkit`](mcpkit) — Сервер Model Context Protocol (`2024-11-05`) поверх STDIO и HTTP/SSE с поддержкой Bearer-токенов, TLS, ограничения частоты запросов, конструктором схем (`NewTool`, `Required`, `Min`, `Max`, `Default`) и обобщёнными адаптерами `TypedHandler[T]` / `TextHandler[T]`.
- [`uiakit`](uiakit) — Движок Windows COM UI Automation (`CUIAutomation8`) на чистом Go без CGO: извлечение деревьев доступности Win32, WPF, WinUI 3, Qt и браузерного `RootWebArea` в компактный Set-of-Marks Markdown-дайджест, а также фоновое выполнение `InvokePattern`, `ValuePattern`, `TogglePattern` и `SelectionItemPattern`.

## Установка

```bash
go get github.com/lemon4ksan/ctrl
```

## Использование

### Маршрутизатор CLI с POSIX-флагами и стилизованным выводом

```go
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/lemon4ksan/ctrl"
)

func main() {
	app := ctrl.NewApp("seal", ctrl.Version, "Утилита архивации и сжатия",
		ctrl.NewCommand("create", "Создать архив из файлов", "seal create [flags] <archive> <files...>",
			func(ctx context.Context, args []string, stdout, stderr io.Writer) error {
				fs := flag.NewFlagSet("create", flag.ContinueOnError)
				level := fs.Int("l", 3, "Уровень сжатия (1-9)")
				verbose := fs.Bool("v", false, "Подробный вывод")

				posArgs, err := ctrl.ParseInterspersedFlags(fs, args)
				if err != nil {
					return err
				}
				if len(posArgs) < 2 {
					return fmt.Errorf("требуется указать путь к архиву и входные файлы")
				}

				fmt.Fprintf(stdout, "%s Создание %s (level=%d, verbose=%v)\n",
					ctrl.BadgePass(), posArgs[0], *level, *verbose)
				return nil
			},
			"c", "add",
		),
	)

	app.CommandGroups = []ctrl.CommandGroup{
		{Title: "Основные операции", Commands: []string{"create"}},
	}

	if err := app.Run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", ctrl.BadgeFail(), err)
		os.Exit(1)
	}
}
```

### Декларативный MCP-сервер (`mcpkit`)

```go
package main

import (
	"context"

	"github.com/lemon4ksan/ctrl/mcpkit"
)

type ExecArgs struct {
	Command   string `json:"command"`
	TimeoutMS int    `json:"timeout_ms"`
}

func main() {
	srv := mcpkit.NewServer("agent-tools", "1.0.0")

	tool := mcpkit.NewTool("exec_command", "Выполнить команду оболочки").
		String("command", "Командная строка", mcpkit.Required()).
		Int("timeout_ms", "Таймаут выполнения в миллисекундах", mcpkit.Min(100), mcpkit.Default(5000)).
		Build()

	srv.RegisterTool(tool, mcpkit.TextHandler(func(ctx context.Context, args ExecArgs) (string, error) {
		return "выполнено: " + args.Command, nil
	}))

	_ = srv.ServeStdio(context.Background())
}
```

## Тестирование и проверка

```bash
# Запуск модульных тестов
make test

# Запуск тестов с детектором гонок
make race

# Запуск бенчмарков и проверка 0 allocs/op
make bench

# Запуск линтеров
make lint
```

## Лицензия

BSD 3-Clause License. Полный текст приведён в файле [LICENSE](LICENSE).
