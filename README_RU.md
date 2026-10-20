<div align="center">

<img src="assets/logo.svg" width="120" height="120" alt="ctrl" />

# ctrl

### Фреймворк CLI, TUI, MCP и UI Automation для Go без внешних зависимостей

_«В терминале задержка и визуальный шум — это состояние по умолчанию. Пусть ctrl станет вашей точной командной поверхностью без единой лишней зависимости»_

[![Go Version](https://img.shields.io/badge/go-1.27%2B-007d9c?logo=go&logoColor=white&style=flat-square)](https://go.dev/)
[![Go Reference](https://img.shields.io/badge/godoc-reference-007d9c?style=flat-square)](https://pkg.go.dev/github.com/lemon4ksan/ctrl)
[![License](https://img.shields.io/badge/license-BSD--3--Clause-blue?style=flat-square)](LICENSE)
[![Zero-Dependencies](https://img.shields.io/badge/dependencies-0%20external-brightgreen?style=flat-square)](#сравнение-возможностей)
[![Zero-Alloc](https://img.shields.io/badge/memory-0%20B%2Fop%20%7C%200%20allocs-success?style=flat-square)](#производительность-и-гарантии-памяти)
[![Ecosystem](https://img.shields.io/badge/ecosystem-foundation-orange?style=flat-square)](https://github.com/lemon4ksan/foundation)
[![POSIX Semantics](https://img.shields.io/badge/POSIX-Flags%20%26%20Clumping-blueviolet?style=flat-square)](argkit)

**ctrl** — фреймворк командной строки, терминального интерфейса (TUI), маршрутизатор Model Context Protocol (`mcpkit`) и движок Windows UI Automation (`uiakit`) для Go. Предоставляет полноценный POSIX-парсинг флагов (склеивание коротких флагов, прикреплённые значения, перемежающиеся аргументы), маршрутизацию подкоманд с логическими группами, автоопределение возможностей TTY, визуальные компоненты (таблицы, рамки, прогресс-бары, спарклайны, бейджи), декларативный конструктор MCP-схем и zero-allocation обход COM-дерева доступности.

#### [English](README.md) • Русский • [Концепция архитектуры](docs/CONCEPT.md)

</div>

## Установка

Требуется Go версии `1.27` или выше.

```bash
go get github.com/lemon4ksan/ctrl
```

## Быстрый старт

Создание утилиты командной строки с подкомандами, POSIX-флагами и стилизованным терминальным выводом:

```go
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/lemon4ksan/ctrl"
	"github.com/lemon4ksan/ctrl/tuikit"
)

func main() {
	app := ctrl.NewApp("seal", ctrl.Version, "Высокопроизводительная утилита сжатия и архивации",
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
					return fmt.Errorf("требуется указать путь к архиву и файлы")
				}

				fmt.Fprintf(stdout, "%s Создание архива %s (уровень %d, verbose=%v)\n",
					tuikit.BadgePass(), posArgs[0], *level, *verbose)
				return nil
			},
			"c", "add",
		),
	)

	app.CommandGroups = []ctrl.CommandGroup{
		{
			Title:    "Основные операции",
			Commands: []string{"create"},
		},
	}

	if err := app.Run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", tuikit.BadgeFail(), err)
		os.Exit(1)
	}
}
```

## Сравнение возможностей

| Возможность | Стандартная библиотека (`flag`) | `spf13/cobra` | `ctrl` |
| :--- | :---: | :---: | :---: |
| **Внешние зависимости** | **0** | ~15+ транзитивных | **0** |
| **Перемежающиеся аргументы (Interspersed)** | ✗ (останавливается на 1-м позиционном) | ✓ | **✓ (`argkit`)** |
| **Склеивание коротких флагов (`-la`)** | ✗ | ✓ | **✓ (`argkit`)** |
| **Прикреплённые значения (`-I*.tmp`)** | ✗ | ✓ | **✓ (`argkit`)** |
| **Нечёткие подсказки опечаток** | ✗ | ✓ | **✓ (Левенштейн)** |
| **Маршрутизация подкоманд** | ✗ | ✓ | **✓ (`tuikit.App`)** |
| **Автодетект возможностей TTY** | ✗ | ✗ (нужны внешние пакеты) | **✓ (`tuikit.ProbeTerminal`)** |
| **Поддержка `NO_COLOR` и пайпов** | ✗ | ✗ | **✓ (Встроено)** |
| **Таблицы, рамки, прогресс-бары** | ✗ | ✗ (требует lipgloss) | **✓ (`tuikit`)** |

## Архитектура пакетов

### 1. [`argkit`](argkit) — Движок аргументов и POSIX-флагов
- **Свободное перемежение**: Позиционные аргументы и флаги могут идти в любом порядке (`app -v src/ -o=bin dst/`).
- **Склеивание флагов (Clumping)**: Комбинация `-la` автоматически разворачивается в `-l` и `-a`.
- **Прикреплённые параметры**: Флаги вида `-I*.tmp` прозрачно транслируются в `-I=*.tmp`.
- **Терминатор POSIX `--`**: Все аргументы после `--` трактуются строго как позиционные.
- **Подсказки опечаток**: Автоматический расчёт расстояния Левенштейна на стековых буферах для предложения правильного флага (`"flag provided but not defined: --timout (did you mean: --timeout?)"`).
- **Нормализация токенов**: Устранение артефактов токенизаторов PowerShell и CMD (`argkit.NormalizeArgs`).

### 2. [`tuikit`](tuikit) — Терминальный интерфейс и маршрутизация CLI
- **Маршрутизатор подкоманд (`tuikit.App`)**: Поддержка алиасов, группировка справки по секциям, команды по умолчанию.
- **Детектор терминала (`tuikit.ProbeTerminal`)**: Определение интерактивности TTY, включение Virtual Terminal на Windows 10/11.
- **Выровненные таблицы (`tuikit.Table`)**: Выравнивание по левому, правому краю или по центру с автоматическим подсчётом ширины колонок.
- **Рамочные карточки (`tuikit.Box`)**: Рамки с прямыми границами (`BorderSingle`, `BorderRounded`, `BorderHeavy`).
- **Живой прогресс (`tuikit.Progress`)**: Скорость (MB/s или шт/с), расчёт ETA, автоматическое переключение на периодический лог в CI/пайпах.
- **Спарклайны (`tuikit.RenderSparkline`)**: Компактные графики трендов (` ▂▃▅▆▇█`).
- **Бейджи и этапы пайплайна**: Статусные метки (`BadgePass`, `BadgeFail`, `BadgeWarn`, `BadgeActive`) и нумерованные шаги (`RenderStep`).

### 3. [`mcpkit`](mcpkit) — Маршрутизатор Model Context Protocol v1.0 и Fluent-конструкторы
- **Транспорты JSON-RPC 2.0 STDIO и HTTP/SSE**: Полноценный сервер MCP `2024-11-05` с буфером сканера 16 МБ, Bearer-аутентификацией, TLS и ограничителем частоты запросов (Token Bucket).
- **Декларативный конструктор схем (`mcpkit.NewTool`)**: Создание JSON Schema определений инструментов с функциональными опциями (`mcpkit.Required()`, `mcpkit.Min()`, `mcpkit.Max()`, `mcpkit.MinItems()`, `mcpkit.Default()`) без вложенных `map[string]any`.
- **Типизированные обработчики**: Обобщённые адаптеры `TypedHandler[T]`, `LenientHandler[T]`, `TextHandler[T]`, `StrictTextHandler[T]` и `NoArgsHandler`.

### 4. [`uiakit`](uiakit) — Windows COM UI Automation (`CUIAutomation8`)
- **Zero-Allocation COM-вызовы**: Прямые вызовы таблиц виртуальных методов `IUIAutomation2` через `foundation/silicon/comkit` (`Call0`–`Call5`) с пакетной предвыборкой свойств `IUIAutomationCacheRequest`.
- **Set-of-Marks и DOM**: Извлечение деревьев доступности Win32, WPF, WinUI 3, Qt и браузерного `RootWebArea` в компактный Markdown-дайджест.
- **Паттерны действий**: Нативное выполнение `IUIAutomationInvokePattern`, `ValuePattern`, `TogglePattern` и `SelectionItemPattern` без перехвата системного курсора.

## Производительность и гарантии памяти

- **`VisibleWidth`**: Подсчёт визуальной ширины моноширинных символов без аллокаций в куче.
- **`Levenshtein`**: Использует стек (`[65]int`, `[64]rune`) для типичных флагов $\le 64$ символов — строго **0 аллокаций/оп**.
- **`LinesSeq`**: Push-итератор строк Go 1.27 без создания промежуточных срезов.

## Разработка и тестирование

```bash
# Запуск модульных тестов
make test

# Запуск тестов с детектором гонок
make race

# Запуск линтеров
make lint

# Запуск бенчмарков
make bench
```

## 🧾 Лицензия

Лицензировано под **BSD 3-Clause License**. Подробности в файле [LICENSE](LICENSE).
