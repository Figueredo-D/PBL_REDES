# 🚗 VaiJunto — Sistema de Caronas Compartilhadas

O **VaiJunto** é um sistema distribuído de caronas compartilhadas desenvolvido em **Go** para a disciplina de Concorrência e Conectividade (PBL 1). O projeto utiliza arquitetura cliente-servidor com comunicação via **TCP Sockets puros**, serialização de dados em **JSON**, concorrência thread-safe com travas e empacotamento completo via **Docker**.

---

## 🛠️ Pré-requisitos do Sistema

Antes de iniciar, certifique-se de ter os seguintes softwares instalados na sua máquina:

1. **Docker Desktop** (versão 20.10 ou superior)
   - [Download para Windows / Mac / Linux](https://www.docker.com/products/docker-desktop/)
2. **Git**
   - [Download Git](https://git-scm.com/)
3. *(Opcional)* **Go 1.22+**
   - Necessário apenas se desejar rodar os executáveis ou o teste de estresse diretamente no ambiente local sem utilizar os containers.

---

## 📂 Estrutura do Repositório

PBL_REDES/
├── cmd/
│   ├── driver-client/      # CLI Interativa do Motorista
│   ├── passenger-client/   # CLI Interativa do Passageiro
│   ├── server/             # Servidor Central de Caronas (TCP)
│   └── stress-test/        # Script de Teste de Concorrência
├── internal/
│   ├── domain/             # Entidades de Domínio (Ride, Segment, Booking)
│   ├── protocol/           # Definição de Mensagens JSON e Ações da API
│   └── repository/         # Armazenamento em Memória Thread-Safe
├── Dockerfile              # Build Multi-Stage (Compilação e Alvos)
├── docker-compose.yml      # Orquestração do Servidor e Clientes
└── README.md               # Documentação do Projeto

---

## 🚀 Passo a Passo para Execução

### Opção A: Execução em Uma Única Máquina (Testes Locais)

Siga estes passos para rodar o servidor e abrir as interfaces do motorista e passageiro na sua própria máquina local:

#### 1. Clonar o Repositório
Abra o terminal e clone o projeto:
git clone https://github.com/Figueredo-D/PBL_REDES.git
cd PBL_REDES

#### 2. Subir o Servidor em Background
Inicie o container do servidor central:

docker-compose up -d --build server

*(Para verificar se o servidor está rodando, execute `docker-compose ps`)*.

#### 3. Iniciar o Cliente Motorista
Em uma janela do terminal, execute:

docker-compose run --rm driver-client

#### 4. Iniciar o Cliente Passageiro
Em **outra janela de terminal separada**, execute:

docker-compose run --rm passenger-client

> **No menu do passageiro:** Busque por viagens na mesma rota/data cadastrada e faça a reserva da vaga.

---

### Opção B: Execução em Múltiplas Máquinas (Rede Local / Laboratório)

Para testar a conectividade em máquinas diferentes (ex: Máquina A rodando o Servidor e Máquina B rodando o Passageiro na mesma rede Wi-Fi/Cabo):

#### 1. Na Máquina A (Onde ficará o Servidor):
Descubra o IP local da máquina:
- **Windows (PowerShell):** `ipconfig` (procure por *Endereço IPv4*, ex: `192.168.1.15`)
- **Linux/Mac:** `ip a` ou `ifconfig`

Em seguida, suba o servidor:

docker-compose up -d --build server

#### 2. Na Máquina B (Cliente Remoto):
Garanta que o repositório foi clonado e execute o cliente informando o IP da **Máquina A** através da variável `SERVER_HOST`:

- **Para abrir o Passageiro:**

docker-compose run --rm -e SERVER_HOST=192.168.1.15 passenger-client

- **Para abrir o Motorista:**

docker-compose run --rm -e SERVER_HOST=192.168.1.15 driver-client

*(Substitua `192.168.1.15` pelo IP real retornado na Máquina A)*.

---

## 🧪 Como Executar o Teste de Concorrência e Estresse

O projeto possui um script de teste para validar a integridade das reservas e evitar o *overbooking* (garantindo atomicidade quando dezenas de passageiros tentam reservar a mesma vaga ao mesmo tempo).

Com o servidor rodando (`docker-compose up -d server`), execute:

go run cmd/stress-test/main.go

*O script dispara 30 requisições simultâneas concorrendo por apenas 3 assentos disponíveis. O sistema deve confirmar exatamente 3 reservas e rejeitar com erro as 27 restantes.*

---

## 🛑 Como Encerrar o Sistema

Para parar o servidor e remover os containers criados pelo Docker Compose:

docker-compose down

---

## ❓ Resolução de Problemas Comuns

- **Erro: `port is already allocated` (Porta 8080 em uso)**
  - Verifique se já existe outra instância do servidor rodando com `docker-compose ps` e execute `docker-compose down`.

- **Erro de Conexão entre Máquinas Diferentes**
  - Certifique-se de que ambas as máquinas estão conectadas na mesma sub-rede.
  - Verifique se o Firewall do sistema operacional na Máquina A não está bloqueando conexões de entrada na porta TCP 8080.