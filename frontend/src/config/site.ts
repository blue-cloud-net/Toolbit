/**
 * 站点级常量。
 *
 * 只放「全站唯一」的事实：仓库地址、版本号、版权年份。
 * 组件里**禁止**再写死这些值。
 */

/** 项目公开源码仓库。 */
export const REPO_URL = 'https://github.com/blue-cloud-net/Toolbit'

/** 版本发布记录。 */
export const RELEASES_URL = `${REPO_URL}/releases`

/** 前端版本号，构建期可由 `VITE_APP_VERSION` 覆盖。 */
export const APP_VERSION: string = import.meta.env.VITE_APP_VERSION ?? '0.2.1'

/** 版权年份（运行时取当年）。 */
export const COPYRIGHT_YEAR = new Date().getFullYear()
