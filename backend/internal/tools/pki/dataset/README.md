# OID 数据集

> 状态：Stable ｜ 适用版本：v0.1.0 ｜ 最后更新：2026-11-03

本目录是 `oid-lookup` 工具的**离线数据集**：运行时**不联网**，全部查询都在本地完成。

## 文件

| 文件 | 内容 |
| --- | --- |
| `oid_nodes.csv` | `oid,name,description,kind,status,source,sourceUrl,aliases` |
| `oid_zh.csv` | `oid,nameZh,descriptionZh` |

`oid_nodes.csv` 与 `oid_zh.csv` 都由后端在启动时读取并合并：前者提供官方英文名与来源，后者提供中文名与中文说明；
缺中文时回退到英文。前端只使用 API 返回的合并结果，不再单独读数据文件。

`kind` 由生成器计算：`root`（`0` / `1` / `2`）、`branch`（有子节点）、`leaf`（叶子）。

## 数据来源

**官方来源**（英文名与来源一律取自官方注册表）：

| 来源 | 覆盖范围 |
| --- | --- |
| [IANA SMI Numbers](https://www.iana.org/assignments/smi-numbers/) | `1.3.6.1.*`（SNMP / PKIX / S/MIME / LTANS） |
| [IANA Security SMIME](https://www.iana.org/assignments/smi-numbers/) | `1.2.840.113549.1.9.16.*` |
| IETF RFC（见各条目 `source` 列） | X.509 扩展、PKCS、X.500 属性、算法 |
| ITU-T X.660 / X.509 / X.520 | 根节点 `0` / `1` / `2` 及其命名 |
| GM/T 0002 / 0003 / 0004 / 0006、GB/T 32918 | `1.2.156.10197.*`（SM2 / SM3 / SM4） |

> 第三方 OID 站点（如 `oid-base.com`）**不**作为数据来源，只在人工核对时参考。

## 覆盖范围与取舍

收录（约 1160 个节点）：

- ITU / ISO 根节点与常用上级分支；
- **PKIX**（`1.3.6.1.5.5.7.*`）：证书扩展、扩展密钥用途、算法、CMC、CMP、SCVP、属性证书；
- **S/MIME**（`1.2.840.113549.1.9.16.*`）：模块、内容类型、属性、算法、PSKC；
- **PKCS**（`1.2.840.113549.1.*`）：RSA / OAEP / PSS、PKCS #7 / #9 / #12、PBKDF2；
- **X.509 扩展**（`2.5.29.*`）与 **X.500 属性**（`2.5.4.*`）；
- **NIST 算法**（`2.16.840.1.101.3.4.*`）、**ANSI X9.62**（`1.2.840.10045.*`）；
- **国密**（`1.2.156.10197.*`）：SM2 / SM3 / SM4 及其模式；
- **SNMP 主分支**（`1.3.6.1.2.1.*` 的顶层分组、`1.3.6.1.6` / `1.3.6.1.7` / `1.3.6.1.8`）。

**不收录**：

- `1.3.6.1.4.1` 下的**私有企业编号（PEN）**：条目数以万计，且不属于标准体系；
- `1.3.6.1.2.1` 下 MIB 的**完整对象树**：仅收录顶层分组，不展开到每个被管对象；
- ITU / ISO 各成员体的完整注册库：没有可批量获取的公开数据。

未收录的 OID **不是错误**：查询返回 `200` 与「最近祖先」，由用户据此判断。

## 刷新流程

源表在 `scripts/oid/` 下：

| 目录 | 内容 |
| --- | --- |
| `curated/*.csv` | IANA 未覆盖分支的完整条目（英文 + 中文） |
| `zh/*.csv` | IANA 已覆盖分支的中文覆盖表（只有中文两列） |

刷新步骤：

```bash
node scripts/oid/gen-oid-dataset.mjs            # 拉取 IANA 注册表并生成数据集
node scripts/oid/gen-oid-dataset.mjs --offline  # 用 scripts/oid/.cache/ 的缓存生成
node scripts/oid/gen-oid-dataset.mjs --check    # 只校验已提交的数据集是否与源表一致
```

生成器会强制：

1. OID 唯一、父节点存在、根节点 `0` / `1` / `2` 齐备；
2. 英文名、英文说明、中文名、中文说明**全部非空**（缺中文即报错，杜绝半中半英）；
3. 状态取值只能是 `current` / `reserved` / `deprecated` / `obsolete`。

内容变化时**必须**同步更新 `backend/internal/tools/pki/oiddataset.go` 里的 `OIDDatasetVersion`。
