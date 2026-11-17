package wellness

import "strings"

// Return operator-facing text instead of leaking database/validation internals.
func localizedError(language string, status int) string {
	language = strings.ToLower(strings.TrimSpace(strings.Split(language, ",")[0]))
	group := 0
	if strings.HasPrefix(language, "zh") {
		group = 1
	}
	if strings.HasPrefix(language, "pt") {
		group = 2
	}
	messages := map[int][3]string{
		400: {"Invalid filters. Check the date range and filter values.", "筛选条件无效，请检查日期范围和筛选值。", "Filtros inválidos. Verifique o intervalo de datas e os valores."},
		404: {"The record no longer exists. Refresh the list.", "记录已不存在，请刷新列表。", "O registro não existe mais. Atualize a lista."},
		500: {"Unable to read wellness data. Please try again later.", "健康数据读取失败，请稍后重试。", "Não foi possível ler os dados de saúde. Tente novamente mais tarde."},
		502: {"The BYND service is unavailable or returned an invalid response.", "BYND 服务暂不可用或返回了无效响应。", "O serviço BYND está indisponível ou retornou uma resposta inválida."},
	}
	if message, ok := messages[status]; ok {
		return message[group]
	}
	return messages[500][group]
}
