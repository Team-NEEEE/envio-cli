# Envio

차세대 영지식 기반 환경변수 및 시크릿 관리 CLI

## 🚀 소개

Envio는 환경변수(.env) 및 시크릿을 안전하게 관리하기 위한 CLI 기반 플랫폼입니다.  
서버가 절대 평문 데이터를 알 수 없는 **Zero-Knowledge 아키텍처**를 기반으로 동작합니다.

:contentReference[oaicite:0]{index=0}

## 🔐 핵심 특징

- Zero-Knowledge 암호화 구조
- Repo 단위 세션 격리
- CLI 기반 간편 워크플로우
- AI 에이전트(MCP) 안전 연동
- 메모리 기반 환경변수 주입 (Zero-Disk)

## ⚙️ 주요 명령어

```bash
envio login     # 로그인 및 키 생성
envio link      # 프로젝트 연결
envio pull      # 환경변수 가져오기
envio push      # 환경변수 업로드
envio run       # 실행 시 환경변수 주입