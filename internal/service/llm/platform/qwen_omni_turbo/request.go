// Copyright 2021 ecodeclub
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package qwen_omni_turbo

type SendRequest struct {
	Model         string        `json:"model"`                    // defaults to "qwen-omni-turbo"
	Messages      []Messages    `json:"messages,omitempty"`       // required, at least one message
	Stream        bool          `json:"stream"`                   // defaults to true
	StreamOptions StreamOptions `json:"stream_options,omitempty"` // optional, defaults to { "include_usage": true }
	Modalities    []string      `json:"modalities,omitempty"`     // optional, defaults to ["text"]
	Audio         Audio         `json:"audio,omitempty"`
}

// SendRequestBuilder 是构建 SendRequest 的 builder
type SendRequestBuilder struct {
	request SendRequest
}

// NewSendRequestBuilder 创建一个新的构建器，设置默认值
func NewSendRequestBuilder(ms []Messages) *SendRequestBuilder {
	return &SendRequestBuilder{
		request: SendRequest{
			Model:  "qwen-omni-turbo",
			Stream: true,
			StreamOptions: StreamOptions{
				IncludeUsage: true,
			},
			Messages:   ms,
			Modalities: []string{"text"},
			Audio:      Audio{Voice: "Cherry", Format: "wav"},
		},
	}
}

// SendRequestOption 选项函数类型
type SendRequestOption func(*SendRequest)

// WithModel 设置模型名称选项
func WithModel(model string) SendRequestOption {
	return func(r *SendRequest) {
		r.Model = model
	}
}

// WithModalities 设置模态选项
func WithModalities(modalities []string) SendRequestOption {
	return func(r *SendRequest) {
		r.Modalities = modalities
	}
}

// WithAudio 设置音频选项
func WithAudio(audio Audio) SendRequestOption {
	return func(r *SendRequest) {
		r.Audio = audio
	}
}

// Build 应用所有选项并构建最终的 SendRequest
func (b *SendRequestBuilder) Build(opts ...SendRequestOption) SendRequest {
	// 创建副本，避免修改原始结构
	result := b.request

	// 应用所有选项
	for _, opt := range opts {
		opt(&result)
	}

	return result
}

type StreamOptions struct {
	IncludeUsage bool `json:"include_usage"` // defaults to true
}

type Audio struct {
	Voice  string `json:"voice"`
	Format string `json:"format"`
}

type Messages struct {
	Role    string    `json:"role"` // required, e.g., "user", "assistant", "system"
	Content []Content `json:"content"`
}

type Content interface {
	GetType() string
}

const (
	ContentTypeText       = "text"
	ContentTypeVideo      = "video"
	ContentTypeImage      = "image_url"
	ContentTypeInputAudio = "input_audio"
)

type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (t TextContent) GetType() string {
	return t.Type
}

func NewTextContent(text string) TextContent {
	return TextContent{
		Type: ContentTypeText,
		Text: text,
	}
}

type VideoContent struct {
	Type  string   `json:"type"`
	Video []string `json:"video"`
}

func (v VideoContent) GetType() string {
	return v.Type
}

func NewVideoContent(video []string) VideoContent {
	return VideoContent{
		Type:  ContentTypeVideo,
		Video: video,
	}
}

type ImageContent struct {
	Type     string   `json:"type"`
	ImageUrl ImageUrl `json:"image_url"`
}

func (i ImageContent) GetType() string {
	return i.Type
}

type ImageUrl struct {
	Url string `json:"url"`
}

func NewImageContent(url string) ImageContent {
	return ImageContent{
		Type:     ContentTypeImage,
		ImageUrl: ImageUrl{Url: url},
	}
}

type InputAudioContent struct {
	Type       string     `json:"type"`
	InputAudio InputAudio `json:"input_audio"`
}

func (i InputAudioContent) GetType() string {
	return i.Type
}

type InputAudio struct {
	Data   string `json:"data"`
	Format string `json:"format"`
}

func NewInputAudioContent(data string, fmt string) InputAudioContent {
	return InputAudioContent{
		Type:       ContentTypeInputAudio,
		InputAudio: InputAudio{Data: data, Format: fmt},
	}
}
