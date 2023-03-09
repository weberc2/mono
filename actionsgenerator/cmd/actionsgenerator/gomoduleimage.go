package main

type GoModuleImage func(module string) *Image

func GoModImage(mainPackage string) GoModuleImage {
	return func(module string) *Image {
		return GoImage(module, mainPackage)
	}
}

func (image GoModuleImage) SetDockerfile(dockerfile string) GoModuleImage {
	return func(module string) *Image {
		return image(module).SetDockerfile(dockerfile)
	}
}

func (image GoModuleImage) SetRegistry(registry *Registry) GoModuleImage {
	return func(module string) *Image {
		return image(module).SetRegistry(registry)
	}
}

func (image GoModuleImage) SetSinglePlatform(platform string) GoModuleImage {
	return func(module string) *Image {
		return image(module).SetSinglePlatform(platform)
	}
}
