# 🚗 VaiJunto — Sistema de Caronas Compartilhadas

> Sistema distribuído de caronas compartilhadas desenvolvido em **Go** para a disciplina de **Concorrência e Conectividade — PBL 1**.

O **VaiJunto** implementa uma arquitetura **cliente-servidor**, permitindo que motoristas disponibilizem viagens e passageiros realizem reservas de vagas.

O projeto utiliza:

* **Go** para desenvolvimento da aplicação
* **TCP Sockets puros** para comunicação entre cliente e servidor
* **JSON** para serialização e troca de mensagens
* **Concorrência thread-safe** com mecanismos de sincronização
* **Docker** para empacotamento e execução
* **Docker Compose** para orquestração dos serviços
* **Teste de estresse** para validação da concorrência e prevenção de *overbooking*

---

## 📌 Funcionalidades

### Motorista

* Cadastro de viagens
* Definição de origem e destino
* Definição de data e horário
* Definição da quantidade de vagas disponíveis
* Gerenciamento das viagens cadastradas

### Passageiro

* Busca por viagens disponíveis
* Consulta de viagens por rota e data
* Reserva de vagas
* Recebimento de confirmação ou erro da reserva

### Servidor

* Gerenciamento centralizado das viagens
* Processamento de múltiplos clientes simultaneamente
* Comunicação através de TCP
* Processamento de mensagens JSON
* Controle de concorrência nas reservas
* Prevenção de *overbooking*

---

## 🛠️ Tecnologias Utilizadas

| Tecnologia         | Utilização                   |
| ------------------ | ---------------------------- |
| **Go**             | Linguagem principal          |
| **TCP Sockets**    | Comunicação cliente-servidor |
| **JSON**           | Serialização das mensagens   |
| **Docker**         | Containerização              |
| **Docker Compose** | Orquestração dos containers  |
| **Git**            | Controle de versão           |

---

## 📂 Estrutura do Projeto

```text
PBL_REDES/
├── cmd/
│   ├── driver-client/       # CLI interativa do motorista
│   ├── passenger-client/    # CLI interativa do passageiro
│   ├── server/              # Servidor central de caronas
│   └── stress-test/         # Teste de concorrência e estresse
│
├── internal/
│   ├── domain/              # Entidades de domínio
│   │   ├── Ride
│   │   ├── Segment
│   │   └── Booking
│   │
│   ├── protocol/            # Mensagens JSON e ações da API
│   │
│   └── repository/           # Armazenamento em memória thread-safe
│
├── Dockerfile                # Build multi-stage
├── docker-compose.yml        # Orquestração dos serviços
└── README.md                 # Documentação do projeto
```

---

# 🚀 Como Executar

## 📋 Pré-requisitos

Antes de executar o projeto, certifique-se de possuir:

### 🐳 Docker Desktop

**Docker Desktop 20.10 ou superior**

[Download do Docker Desktop](https://www.docker.com/products/docker-desktop/?utm_source=chatgpt.com)

### 🔧 Git

[Download do Git](https://git-scm.com/?utm_source=chatgpt.com)

### 🐹 Go 1.22+

Opcional.

O Go é necessário apenas caso você queira executar diretamente os executáveis ou o teste de estresse fora dos containers Docker.

---

# 💻 Opção A — Execução em Uma Única Máquina

Esta opção permite executar o servidor, o cliente motorista e o cliente passageiro na mesma máquina.

## 1. Clonar o repositório

Abra um terminal e execute:

```bash
git clone https://github.com/Figueredo-D/PBL_REDES.git
cd PBL_REDES
```

---

## 2. Iniciar o servidor

Execute:

```bash
docker-compose up -d --build server
```

Para verificar se o servidor está funcionando:

```bash
docker-compose ps
```

O servidor ficará disponível na porta:

```text
TCP 8080
```

---

## 3. Iniciar o cliente motorista

Abra um novo terminal dentro da pasta do projeto:

```bash
docker-compose run --rm driver-client
```

A interface do motorista será iniciada no terminal.

---

## 4. Iniciar o cliente passageiro

Abra **outro terminal separado** e execute:

```bash
docker-compose run --rm passenger-client
```

No menu do passageiro:

1. Busque por viagens;
2. Informe a rota desejada;
3. Informe a data;
4. Selecione uma viagem disponível;
5. Realize a reserva da vaga.

> **Importante:** utilize uma rota e uma data que tenham sido previamente cadastradas pelo motorista.

---

# 🌐 Opção B — Execução em Múltiplas Máquinas

Também é possível executar o sistema em computadores diferentes conectados à mesma rede local.

### Exemplo

```text
┌──────────────────────┐
│      Máquina A       │
│                      │
│   VaiJunto Server    │
│      TCP :8080       │
└──────────┬───────────┘
           │
        Rede local
           │
┌──────────▼───────────┐
│      Máquina B       │
│                      │
│ Passenger / Driver   │
└──────────────────────┘
```

---

## 1. Descobrir o IP da Máquina A

A Máquina A será responsável por executar o servidor.

### Windows

No PowerShell ou Prompt de Comando:

```powershell
ipconfig
```

Procure pelo campo:

```text
Endereço IPv4
```

Por exemplo:

```text
192.168.1.15
```

### Linux / macOS

Execute:

```bash
ip a
```

ou:

```bash
ifconfig
```

---

## 2. Iniciar o servidor na Máquina A

Na Máquina A:

```bash
docker-compose up -d --build server
```

O servidor estará disponível na porta:

```text
8080
```

---

## 3. Conectar um cliente remoto

Na Máquina B, clone o repositório:

```bash
git clone https://github.com/Figueredo-D/PBL_REDES.git
cd PBL_REDES
```

Depois, informe o endereço IP da Máquina A através da variável `SERVER_HOST`.

### Passageiro

```bash
docker-compose run --rm -e SERVER_HOST=192.168.1.15 passenger-client
```

### Motorista

```bash
docker-compose run --rm -e SERVER_HOST=192.168.1.15 driver-client
```

Substitua:

```text
192.168.1.15
```

pelo endereço IPv4 real da Máquina A.

> **Atenção:** as máquinas precisam estar conectadas à mesma rede e a porta TCP **8080** deve estar liberada no firewall da Máquina A.

---

# 🧪 Teste de Concorrência e Estresse

O projeto possui um teste específico para validar o comportamento do sistema diante de múltiplas requisições simultâneas.

O principal objetivo é garantir que duas ou mais requisições não consigam reservar a mesma vaga simultaneamente, evitando o problema conhecido como **overbooking**.

### Cenário do teste

```text
30 passageiros
      │
      │ requisições simultâneas
      ▼
┌─────────────────┐
│     Servidor    │
│                 │
│  3 vagas        │
└────────┬────────┘
         │
    ┌────┴────┐
    ▼    ▼    ▼
   ✓    ✓    ✓     → 3 reservas confirmadas
         
   ✗ ✗ ✗ ✗ ...     → 27 reservas rejeitadas
```

O teste dispara **30 requisições simultâneas** concorrendo por apenas **3 assentos disponíveis**.

O resultado esperado é:

| Resultado            | Quantidade |
| -------------------- | ---------: |
| Reservas confirmadas |      **3** |
| Reservas rejeitadas  |     **27** |
| Total de requisições |     **30** |

---

## ▶️ Executando o teste

Primeiro, certifique-se de que o servidor está funcionando:

```bash
docker-compose up -d server
```

Depois, execute o teste:

```bash
go run cmd/stress-test/main.go
```

O teste verifica se o controle de concorrência do servidor mantém a integridade das reservas mesmo quando várias requisições são processadas simultaneamente.

---

# 🧵 Controle de Concorrência

O armazenamento das informações é realizado em memória e possui mecanismos de sincronização para permitir acesso concorrente seguro.

Durante uma reserva, o sistema precisa garantir que a verificação da disponibilidade e a redução do número de vagas ocorram de maneira **atômica**.

Dessa forma, quando vários passageiros tentam reservar a mesma viagem simultaneamente:

```text
Verificar vaga
      ↓
Existe vaga?
   ↙       ↘
 SIM       NÃO
  ↓         ↓
Reservar   Rejeitar
  ↓
Atualizar vagas
```

O mecanismo de sincronização impede que duas requisições obtenham a mesma vaga disponível.

---

# 🛑 Encerrando o Sistema

Para parar o servidor e remover os containers criados pelo Docker Compose:

```bash
docker-compose down
```

Caso queira reconstruir as imagens posteriormente:

```bash
docker-compose up -d --build
```

---

# ❓ Solução de Problemas

## `port is already allocated`

Esse erro normalmente indica que a porta **8080** já está sendo utilizada por outro processo ou container.

Verifique os containers:

```bash
docker-compose ps
```

Caso exista uma instância anterior do projeto, execute:

```bash
docker-compose down
```

Depois, tente iniciar novamente:

```bash
docker-compose up -d --build server
```

---

## Erro de conexão entre máquinas

Caso um cliente não consiga se conectar ao servidor em outra máquina, verifique:

* As duas máquinas estão conectadas à mesma rede;
* O IP informado em `SERVER_HOST` está correto;
* O servidor está em execução;
* A porta TCP **8080** está disponível;
* O firewall da Máquina A permite conexões de entrada na porta **8080**.

Você pode verificar o endereço IP novamente com:

```powershell
ipconfig
```

no Windows, ou:

```bash
ip a
```

no Linux.

---

# 👥 Arquitetura

O VaiJunto segue uma arquitetura cliente-servidor:

```text
             ┌─────────────────────┐
             │       SERVER        │
             │                     │
             │  TCP :8080          │
             │                     │
             │  Repository         │
             │  Protocol           │
             │  Domain             │
             └──────────┬──────────┘
                        │
              TCP + JSON│
          ┌─────────────┴─────────────┐
          │                           │
┌─────────▼─────────┐       ┌─────────▼─────────┐
│ Driver Client     │       │ Passenger Client  │
│                   │       │                   │
│ CLI               │       │ CLI               │
│                   │       │                   │
│ Cadastrar viagens │       │ Buscar viagens    │
│                   │       │ Reservar vagas    │
└───────────────────┘       └───────────────────┘
```

---

# 📚 Contexto Acadêmico

**Projeto:** VaiJunto — Sistema de Caronas Compartilhadas
**Disciplina:** Concorrência e Conectividade
**Atividade:** PBL 1
**Linguagem:** Go
**Comunicação:** TCP Sockets
**Serialização:** JSON
**Containerização:** Docker

---

# 📄 Licença

Este projeto foi desenvolvido para fins **acadêmicos** no contexto da disciplina de Concorrência e Conectividade.

---

<div align="center">

**🚗 VaiJunto**

*Conectando pessoas, compartilhando caminhos.*

</div>
