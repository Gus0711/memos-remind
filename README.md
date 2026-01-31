# Memos

<img align="right" height="96px" src="https://raw.githubusercontent.com/usememos/.github/refs/heads/main/assets/logo-rounded.png" alt="Memos" />

An open-source, self-hosted note-taking service. Your thoughts, your data, your control — no tracking, no ads, no subscription fees.

[![Home](https://img.shields.io/badge/🏠-usememos.com-blue?style=flat-square)](https://usememos.com)
[![Live Demo](https://img.shields.io/badge/✨-Try%20Demo-orange?style=flat-square)](https://demo.usememos.com/)
[![Docs](https://img.shields.io/badge/📚-Documentation-green?style=flat-square)](https://usememos.com/docs)
[![Discord](https://img.shields.io/badge/💬-Discord-5865f2?style=flat-square&logo=discord&logoColor=white)](https://discord.gg/tfPJa4UmAv)
[![Docker Pulls](https://img.shields.io/docker/pulls/neosmemo/memos?style=flat-square&logo=docker)](https://hub.docker.com/r/neosmemo/memos)

<img src="https://raw.githubusercontent.com/usememos/.github/refs/heads/main/assets/demo.png" alt="Memos Demo Screenshot" height="512" />

### 💎 Featured Sponsors

[**Warp** — The AI-powered terminal built for speed and collaboration](https://go.warp.dev/memos)

<a href="https://go.warp.dev/memos" target="_blank" rel="noopener">
  <img src="https://raw.githubusercontent.com/warpdotdev/brand-assets/main/Github/Sponsor/Warp-Github-LG-02.png" alt="Warp - The AI-powered terminal built for speed and collaboration" width="512" />
</a>

---

[**TestMu AI** - The world's first full-stack Agentic AI Quality Engineering platform](https://www.testmu.ai/?utm_source=memos&utm_medium=sponsor)

<a href="https://www.testmu.ai/?utm_source=memos&utm_medium=sponsor" target="_blank" rel="noopener">
  <img src="https://usememos.com/sponsors/testmu.svg" alt="TestMu AI" height="36" />
</a>

## Overview

Memos is a privacy-first, self-hosted knowledge base that works seamlessly for personal notes, team wikis, and knowledge management. Built with Go and React, it offers lightning-fast performance without compromising on features or usability.

**Why choose Memos over cloud services?**

| Feature           | Memos                          | Cloud Services                |
| ----------------- | ------------------------------ | ----------------------------- |
| **Privacy**       | ✅ Self-hosted, zero telemetry | ❌ Your data on their servers |
| **Cost**          | ✅ Free forever, MIT license   | ❌ Subscription fees          |
| **Performance**   | ✅ Instant load, no latency    | ⚠️ Depends on internet        |
| **Ownership**     | ✅ Full control & export       | ❌ Vendor lock-in             |
| **API Access**    | ✅ Full REST + gRPC APIs       | ⚠️ Limited or paid            |
| **Customization** | ✅ Open source, forkable       | ❌ Closed ecosystem           |

## Features

- **🔒 Privacy-First Architecture**

  - Self-hosted on your infrastructure with zero telemetry
  - Complete data ownership and export capabilities
  - No tracking, no ads, no vendor lock-in

- **📝 Markdown Native**

  - Full markdown support
  - Plain text storage — take your data anywhere

- **⚡ Blazing Fast**

  - Built with Go backend and React frontend
  - Optimized for performance at any scale

- **🐳 Simple Deployment**

  - One-line Docker installation
  - Supports SQLite, MySQL, and PostgreSQL

- **🔗 Developer-Friendly**

  - Full REST and gRPC APIs
  - Easy integration with existing workflows

- **🎨 Beautiful Interface**

  - Clean, minimal design and dark mode support
  - Mobile-responsive layout

- **🔔 Memo Reminders & Web Push Notifications** *(New)*

  - Set date/time reminders on any memo
  - Receive browser and mobile push notifications even when the tab is closed
  - VAPID keys auto-generated on first startup — zero configuration needed
  - Per-device push subscription management
  - Background scheduler checks due reminders every 30 seconds
  - Works on Chrome desktop (WNS) and Chrome Android (FCM)

## Quick Start

### Docker (Recommended)

```bash
docker run -d \
  --name memos \
  -p 5230:5230 \
  -v ~/.memos:/var/opt/memos \
  neosmemo/memos:stable
```

Open `http://localhost:5230` and start writing!

### Try the Live Demo

Don't want to install yet? Try our [live demo](https://demo.usememos.com/) first!

### Other Installation Methods

- **Docker Compose** - Recommended for production deployments
- **Pre-built Binaries** - Available for Linux, macOS, and Windows
- **Kubernetes** - Helm charts and manifests available
- **Build from Source** - For development and customization

See our [installation guide](https://usememos.com/docs/installation) for detailed instructions.

## Reminders & Push Notifications

### How it works

1. **Set a reminder** — Click the bell icon on any memo and pick a date and time.
2. **Enable push notifications** — Go to *Settings → Push Notifications* and click *Enable*. Your browser will ask for notification permission.
3. **Receive notifications** — When a reminder is due, you get a push notification on every device where you enabled it, even if the tab is closed.

### Technical details

The reminder system adds two database tables (`reminder` and `push_subscription`) and seven new API endpoints via Connect RPC:

| Endpoint | Description |
| --- | --- |
| `ReminderService/CreateReminder` | Create a reminder on a memo |
| `ReminderService/ListReminders` | List reminders for the current user |
| `ReminderService/DeleteReminder` | Delete a reminder |
| `UserService/CreateUserPushSubscription` | Register a device for push |
| `UserService/DeleteUserPushSubscription` | Unregister a device |
| `UserService/ListUserPushSubscriptions` | List registered devices |
| `UserService/GetVAPIDPublicKey` | Get the server's VAPID public key |

VAPID keys are generated automatically on first startup and stored in system settings. No manual configuration is required. Push delivery uses the [Web Push protocol](https://web.dev/articles/push-notifications-overview) via the `webpush-go` library.

### Browser support

| Browser | Desktop | Mobile |
| --- | --- | --- |
| Chrome | ✅ | ✅ |
| Edge | ✅ | ✅ |
| Firefox | ✅ | ✅ |
| Safari | ✅ (macOS 13+) | ✅ (iOS 16.4+) |

> **Note:** Push notifications do not work in incognito/private browsing mode.

## Contributing

We welcome contributions of all kinds! Whether you're fixing bugs, adding features, improving documentation, or helping with translations — every contribution matters.

**Ways to contribute:**

- 🐛 [Report bugs](https://github.com/usememos/memos/issues/new?template=bug_report.md)
- 💡 [Suggest features](https://github.com/usememos/memos/issues/new?template=feature_request.md)
- 🔧 [Submit pull requests](https://github.com/usememos/memos/pulls)
- 📖 [Improve documentation](https://github.com/usememos/memos/tree/main/docs)
- 🌍 [Help with translations](https://github.com/usememos/memos/tree/main/web/src/locales)

## Sponsors

Love Memos? [Sponsor us on GitHub](https://github.com/sponsors/usememos) to help keep the project growing!

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=usememos/memos&type=Date)](https://star-history.com/#usememos/memos&Date)

## License

Memos is open-source software licensed under the [MIT License](LICENSE).

## Privacy Policy

Memos is built with privacy as a core principle. As a self-hosted application, all your data stays on your infrastructure. There is no telemetry, no tracking, and no data collection. See our [Privacy Policy](https://usememos.com/privacy) for details.

---

**[Website](https://usememos.com)** • **[Documentation](https://usememos.com/docs)** • **[Demo](https://demo.usememos.com/)** • **[Discord](https://discord.gg/tfPJa4UmAv)** • **[X/Twitter](https://x.com/usememos)**

<a href="https://vercel.com/oss">
  <img alt="Vercel OSS Program" src="https://vercel.com/oss/program-badge.svg" />
</a>