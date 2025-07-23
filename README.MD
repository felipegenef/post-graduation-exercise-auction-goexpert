# Auction API - Golang

**Versão em Português (English version below)**

## 📌 Descrição

Este projeto é uma API desenvolvida como parte de um exercício de pós-graduação em Golang. A API implementa um sistema de leilão onde é possível cadastrar leilões e registrar ofertas (bids) para cada item leiloado. Os dados são armazenados em um banco MongoDB.

Entre os recursos implementados estão:

- Suporte à concorrência com goroutines, mutexes e channels;
- Variáveis de ambiente configuráveis, como:
  - URL do banco de dados;
  - Tamanho do batch de ofertas a serem processadas;
  - Intervalo de tempo entre ofertas;
  - Tempo de duração do leilão (`AUCTION_INTERVAL`);
- Fechamento automático e manual de leilões com base em tempo ou sinais enviados por channel.

> Os testes automatizados se concentram na funcionalidade de fechamento automático e manual dos leilões, conforme requerido no exercício.

---

## ✅ Requisitos do Exercício

Objetivo: adicionar uma funcionalidade de **fechamento automático** de leilões com base em tempo.

### Funcionalidades solicitadas:

- Uma função que calcule o tempo de expiração de um leilão com base nas variáveis de ambiente;
- Uma goroutine que monitore leilões ativos e os feche quando expirados;
- Testes automatizados validando essa funcionalidade.

> Dica: concentre-se principalmente no arquivo `internal/infra/database/auction/create_auction.go`. Lembre-se de tratar corretamente as condições de concorrência.

---

## 🧪 Como Testar a API

### 1. Subir banco de dados de teste

Utilize o docker-compose para subir o MongoDB local:

Comando: docker compose up mongodb

### 2. Executar os testes

Comando: go test ./internal/infra/database/auction -v

---

## 🔬 Explicação dos Testes

### ✅ TestAuctionAutoClose

Verifica se o leilão é fechado automaticamente após o tempo definido em `AUCTION_INTERVAL`.

- Cria um leilão com tempo de expiração definido em 2 segundos;
- Executa duas goroutines simultâneas:
  - A primeira, após 1 segundo, verifica se o leilão ainda está aberto;
  - A segunda, após 3 segundos, verifica se o leilão foi fechado corretamente.

### ✅ TestAuctionManualClose

Valida o encerramento **manual** de um leilão via channel.

- Cria um leilão e armazena seu controle no repositório com mutexes;
- Realiza uma verificação inicial para garantir que está aberto;
- Envia um sinal via channel para forçar o fechamento;
- Após 1 segundo, verifica se o leilão foi corretamente encerrado no banco de dados.

---

# 🏁 English Version

## 📌 Description

This project is an API built as part of a postgraduate exercise in Golang. It implements an auction system where you can register auctions and submit bids for each auctioned item. All data is stored in a MongoDB database.

Key features include:

- Concurrency handling using goroutines, mutexes, and channels;
- Environment variables for:
  - MongoDB connection URL;
  - Bid batch size;
  - Bid insertion interval;
  - Auction duration (`AUCTION_INTERVAL`);
- Automatic and manual auction closing mechanisms.

> Automated tests were added to ensure correct implementation of auction closing features.

---

## ✅ Exercise Requirements

Goal: Implement an **automatic auction closing** feature based on a predefined time.

### Required features:

- A function that calculates auction duration using environment variables;
- A goroutine that checks for expired auctions and updates their status;
- Unit tests validating the automatic closure behavior.

> Tip: focus on `internal/infra/database/auction/create_auction.go`. Pay close attention to concurrency control.

---

## 🧪 How to Test the API

### 1. Start MongoDB test database

Use docker-compose to start the local MongoDB instance:

Command: docker compose up mongodb

### 2. Run the tests

Command: go test ./internal/infra/database/auction -v

---

## 🔬 Test Explanations

### ✅ TestAuctionAutoClose

Tests automatic auction closing using the `AUCTION_INTERVAL` environment variable.

- Creates an auction with a 2-second duration;
- Runs two goroutines:
  - The first waits 1 second and checks if the auction is still open;
  - The second waits 3 seconds and checks that the auction is now closed.

### ✅ TestAuctionManualClose

Tests **manual** auction closing using a channel.

- Creates an auction and registers control metadata (including a close channel) in a map with mutex protection;
- First, verifies that the auction is open;
- Then sends a signal to the auction’s close channel;
- After a 1-second wait, verifies that the auction has been properly closed in the database.

---
