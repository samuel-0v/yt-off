# Contexto do Projeto — yt-off

## Visão geral

O **yt-off** é uma plataforma local e containerizada para gerenciamento de downloads de vídeos e áudios utilizando **yt-dlp** e **FFmpeg**.

O objetivo do projeto é criar uma aplicação que rode totalmente na máquina do usuário, utilizando containers Docker para garantir isolamento, compatibilidade entre sistemas operacionais e facilidade de instalação.

A aplicação foi pensada para uso pessoal em rede local, permitindo que o usuário execute o serviço em seu computador e acesse pelo próprio navegador ou por dispositivos conectados à mesma rede, como celulares.

---

# Motivação

O projeto nasceu da ideia de transformar o uso do `yt-dlp` em uma aplicação completa:

Em vez de executar comandos manualmente:

```bash
yt-dlp URL
```

o usuário terá:

```text
Interface Web
      |
      |
Escolhe vídeo
      |
      |
Seleciona qualidade
      |
      |
Acompanha download
      |
      |
Gerencia arquivos
```

---

# Arquitetura geral

O projeto utiliza uma arquitetura baseada em containers:

```text
                    Usuário

                       |
                       |

                 Frontend Web

                       |
                       |

                 Backend API

                       |
                       |

                 Docker SDK

                       |
                       |

              Container yt-dlp

                       |
                       |

                Volume Downloads
```

---

# Tecnologias

## Frontend

Tecnologias:

* React
* Vite
* TypeScript
* Tailwind CSS
* React Router

Responsabilidades:

* Interface do usuário;
* Consulta de vídeos;
* Seleção de formatos;
* Controle dos downloads;
* Gerenciamento de arquivos.

---

## Backend

Tecnologias:

* Go
* Fiber
* Docker SDK
* SQLite

Responsabilidades:

* API REST;
* Controle de downloads;
* Comunicação com Docker;
* Persistência;
* Gerenciamento de arquivos.

---

## Containers

Serviços:

```text
frontend
backend
yt-dlp
```

O container `yt-dlp` possui:

* yt-dlp;
* FFmpeg.

O backend não executa yt-dlp diretamente no host.

Todo processamento ocorre dentro do container.

---

# Estrutura atual

```text
yt-off/

├── frontend/
│
├── backend/
│
├── docker/
│   └── yt-dlp/
│
├── downloads/
│
├── docker-compose.yml
│
└── README.md
```

---

# Estado atual do desenvolvimento

## Fase 1 — Fundação ✅

Implementado:

* Estrutura Docker Compose;
* Frontend React;
* Backend Go;
* Container yt-dlp;
* Comunicação entre serviços.

Arquitetura inicial funcionando.

---

# Fase 2 — Consulta e download ✅

## Consulta de formatos

Criada API:

```http
POST /api/formats
```

Responsável por:

* consultar informações do vídeo;
* obter formatos disponíveis;
* combinar streams separados.

Exemplo:

Entrada yt-dlp:

```text
137
1080p vídeo

140
áudio AAC
```

Processado:

```json
{
 "label":"1080p MP4",
 "format_id":"137+140"
}
```

---

## Downloads

Criada API:

```http
POST /api/downloads
```

Fluxo:

```text
URL
 |
 |
Formato escolhido
 |
 |
yt-dlp container
 |
 |
Arquivo salvo
```

---

## Status

Implementado:

```text
queued
running
completed
failed
```

---

# Fase 3 — Progresso em tempo real ✅

O yt-dlp passou a executar com:

```bash
--newline
```

O backend interpreta:

```text
[download] 50% of 100MiB at 5MiB/s ETA 00:10
```

Transformando em:

```json
{
 "progress":50,
 "speed":"5MiB/s",
 "eta":"00:10"
}
```

---

# Fase 4 — Persistência SQLite ✅

Antes:

```text
Download
 |
 |
Memória RAM
```

Depois:

```text
Download
 |
 |
SQLite
```

Implementado:

* Banco SQLite;
* Migrations;
* Repository Pattern;
* Persistência após restart.

Tabela:

```text
downloads

id
url
format_id
status
progress
speed
eta
filename
file_size
extension
created_at
updated_at
```

---

# Histórico de downloads ✅

Criado:

```http
GET /api/downloads
```

Retorna:

* downloads realizados;
* status;
* arquivo;
* tamanho;
* datas.

---

# Gerenciamento de arquivos ✅

Implementado:

## Listar arquivos

```http
GET /api/files
```

---

## Baixar arquivo

```http
GET /api/files/:name
```

Suporta:

* download pelo navegador;
* Content-Disposition.

---

## Excluir arquivo

```http
DELETE /api/files/:name
```

Com proteção contra:

```text
../../arquivo
```

---

# Frontend atual

Implementado:

## Home

Fluxo:

```text
Inserir URL

↓

Analisar vídeo

↓

Escolher qualidade

↓

Iniciar download

↓

Acompanhar progresso
```

---

## Downloads

Criada página:

```text
/downloads
```

Permite:

* visualizar histórico;
* ver arquivos;
* baixar arquivos;
* excluir arquivos.

---

# Estado atual das APIs

## Saúde

```http
GET /health
```

---

## Versão

```http
GET /api/version
```

---

## Formatos

```http
POST /api/formats
```

---

## Criar download

```http
POST /api/downloads
```

---

## Consultar download

```http
GET /api/downloads/:id
```

---

## Histórico

```http
GET /api/downloads
```

---

## Arquivos

```http
GET /api/files
```

```http
GET /api/files/:name
```

```http
DELETE /api/files/:name
```

---

# Próxima fase em desenvolvimento

## Fase 5 — Controle de downloads 🚧

Objetivo:

Transformar downloads em tarefas totalmente controláveis.

Implementar:

* cancelamento;
* fila;
* limite de downloads simultâneos;
* controle de processos.

---

# Próximas fases planejadas

## Fase 6 — Rede local

Objetivo:

Permitir acesso por outros dispositivos.

Recursos:

* domínio local;
* mDNS;
* QR Code;
* acesso pelo celular.

Exemplo:

```text
http://yt-off.local
```

---

## Fase 7 — Melhorias de produto

Planejado:

* configurações;
* temas;
* estatísticas;
* organização de downloads;
* favoritos;
* melhorias de UX.

---

# Princípios do projeto

## Local first

O usuário controla seus dados.

Não depende de:

* servidores externos;
* contas;
* serviços terceiros.

---

## Container first

Todas as dependências ficam isoladas.

Benefícios:

* mesma experiência em Windows/Linux/macOS;
* instalação simples;
* ambiente previsível.

---

## API first

Frontend é apenas uma camada visual.

Toda regra de negócio permanece no backend.

---

# Roadmap

```text
✅ Fundação Docker
✅ Backend API
✅ Frontend inicial
✅ Integração yt-dlp
✅ Consulta de formatos
✅ Download real
✅ Progresso
✅ SQLite
✅ Histórico
✅ Gerenciamento de arquivos

🚧 Cancelamento
🚧 Fila
🚧 Controle de concorrência

⬜ Rede local
⬜ Domínio interno
⬜ QR Code
⬜ Mobile
⬜ Configurações
```

---

# Objetivo final

Transformar o yt-off em uma plataforma local completa para gerenciamento de mídia:

```text
Descobrir vídeo
      |
Escolher qualidade
      |
Baixar
      |
Acompanhar
      |
Organizar
      |
Acessar de qualquer dispositivo da rede local
```

O projeto está evoluindo de um simples wrapper do yt-dlp para uma aplicação completa, modular e preparada para uso diário.
