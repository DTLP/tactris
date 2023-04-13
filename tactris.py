import pygame
import random
import copy
import time

# Define constants
BLACK = (0, 0, 0)
WHITE = (255, 255, 255)
GREY = (100, 100, 150)
LGREY = (215, 215, 230)
WIDTH = 40
HEIGHT = 40
MARGIN = 1

# Define all types of blocks
blocks = [
        [[0, 0], [1, 0], [2, 0], [3, 0]],  # I shape
		[[0, 0], [0, 1], [0, 2], [0, 3]],  # I horizontal shape
        [[0, 0], [0, 1], [0, 2], [1, 2]],  # L shape
		[[0, 0], [1, 0], [2, 0], [0, 1]],  # L horizontal shape
        [[0, 0], [0, 1], [0, 2], [1, 0]],  # J shape
		[[0, 0], [1, 0], [2, 0], [2, 1]],  # J horizontal shape
        [[0, 0], [0, 1], [1, 1], [1, 0]],  # O shape
        [[0, 0], [0, 1], [1, 1], [1, 2]],  # S shape
        [[0, 1], [0, 2], [1, 0], [1, 1]],  # Z shape
		[[0, 0], [0, 1], [0, 2], [1, 1]],  # T shape
    ]

def build_grid():
    # Grid containing blocks the player has placed
    # The function generates an empty grid
    global grid, prev_grid

    grid = []
    for row in range(10):
        grid.append([(255, 255, 255)] * 10)
    prev_grid = copy.deepcopy(grid)


def build_ghost_grid():
    # Grid showing where the current_block will be placed on mouse click
    # The function generates an empty grid
    global ghost_grid

    ghost_grid = []
    for row in range(10):
        ghost_grid.append([(255, 255, 255)] * 10)

def generate_block():
    # When called, returns a random block

    return random.choice(blocks)

def can_place_block(block, position, grid):
    # Check if a block can be placed at the mouse cursor

    for b in block:
        row = position[0] + b[0]
        col = position[1] + b[1]
        if row < 0 or row >= 10 or col < 0 or col >= 10:
            return False
        if grid[row][col] == 1:
            return False
    return True

def add_block(block, position, grid):
    # Place current_block at the mouse cursor
    global prev_grid, prev_score

    prev_grid = copy.deepcopy(grid)
    prev_score = copy.deepcopy(score)

    for b in block:
        row = position[0] + b[0]
        col = position[1] + b[1]
        grid[row][col] = 1

def upd_ghost_grid(current_block, position):
    # Draw current_block at the mouse cursor
    # This helps player seeing where the block will be place on mouse click
    global ghost_grid

    build_ghost_grid()
    for b in current_block:
        row = position[0] + b[0]
        col = position[1] + b[1]
        ghost_grid[row][col] = 1


def upd_screen():
    # Refresh the screen to reflect any changes

    for row in range(10):
        for col in range(10):
            color = WHITE
            if ghost_grid[row][col] == 1:
                color = LGREY
            if grid[row][col] == 1:
                color = GREY
            board_cell = pygame.Rect((MARGIN + WIDTH) * col + MARGIN, \
                (MARGIN + HEIGHT) * row + MARGIN, WIDTH, HEIGHT)
            pygame.draw.rect(screen, color, board_cell)


def draw_ghost_block(block, position, grid, ghost_grid):
    # Check if mouse cursor has been moved to another grid cell
    # Moves the ghost_block when mouse is moved
    global prev_position

    try:
        if position != prev_position:
            upd_ghost_grid(current_block, position)

            # Update the grid
            upd_screen()

            # Update the screen
            pygame.display.flip()

        prev_position = copy.deepcopy(position)
    except IndexError:
        pass
    

def clear_lines(grid):
    # Clear lines when they are fully filled in
    global score

    full_rows = []
    for row in range(len(grid)):
        is_full = True
        for col in range(len(grid)):
            if grid[row][col] == (255, 255, 255):
                is_full = False
                break
        if is_full:
            full_rows.append(row)

    full_cols = []
    for col in range(len(grid[0])):
        is_full = True
        for row in range(len(grid)):
            if grid[row][col] == (255, 255, 255):
                is_full = False
                break
        if is_full:
            full_cols.append(col)

    for row in full_rows:
        grid.pop(row)
        grid.insert(0, [(255, 255, 255)] * len(grid[0]))
        score += 10

    for col in full_cols:
        for row in range(len(grid)):
            del grid[row][col]
        for row in range(len(grid)):
            grid[row].insert(0, (255, 255, 255))
        score += 10

def reset_game():
    # Reset the board, score and generate new blocks
    global score, prev_score, grid, current_block, prev_block, next_block

    print("Game over. Score: " + str(score))
	
    current_block = generate_block()
    prev_block = copy.deepcopy(current_block)
    next_block = generate_block()
    score = 0
    prev_score = 0

    build_grid()
		
    pygame.display.flip()

def undo_step():
    # Remove previously placed block
    global prev_grid, grid, score, current_block, prev_block

    grid = copy.deepcopy(prev_grid)
    score = copy.deepcopy(prev_score)
    current_block = copy.deepcopy(prev_block)

    pygame.display.flip()

# Initialize Pygame
pygame.init()

# Set the height and width of the screen
WINDOW_SIZE = [(MARGIN + WIDTH) * 20 + MARGIN, (MARGIN + HEIGHT) * 10 + MARGIN]
screen = pygame.display.set_mode(WINDOW_SIZE)

# Set the title of the window
pygame.display.set_caption("Tactris")

# Initialize the grid
build_grid()
build_ghost_grid()

# Initialize the current block and position
current_block = generate_block()
prev_block = copy.deepcopy(current_block)
next_block = generate_block()
current_position = [1, 10]
next_position = [1, 15]

# Initialize the score
score = 0
prev_score = copy.deepcopy(score)

# Set up buttons
font = pygame.font.Font(None, 36)
font_size = 24
font = pygame.font.SysFont(None, font_size)
text_color = (255, 255, 255)
button_width = 140
button_height = 60
button_color = (0, 255, 0)
# geme reset button
reset_button_x = WINDOW_SIZE[0] - 200
reset_button_y = WINDOW_SIZE[1] - 80
reset_button_text = "RESET"
reset_button_rect = pygame.Rect(
                    reset_button_x, reset_button_y, button_width, button_height)
# undo step button
undo_button_x = WINDOW_SIZE[0] - 350
undo_button_y = WINDOW_SIZE[1] - 80
undo_button_text = "UNDO"
undo_button_rect = pygame.Rect(
                   undo_button_x, undo_button_y, button_width, button_height)

# Loop until the user clicks the close button
done = False

# Set the clock
clock = pygame.time.Clock()

# Previous cursor position
prev_position = []

# Loop until the game is over
while not done:
    # Draw the grid
    screen.fill(BLACK)
    for row in range(10):
        for col in range(10):
            color = WHITE
            if ghost_grid[row][col] == 1:
                color = LGREY
            if grid[row][col] == 1:
                color = GREY
            board_cell = pygame.Rect((MARGIN + WIDTH) * col + MARGIN, \
                (MARGIN + HEIGHT) * row + MARGIN, WIDTH, HEIGHT)
            pygame.draw.rect(screen, color, board_cell)

    # Draw the current block
    cur_text = font.render("Current: ", True, WHITE)
    screen.blit(cur_text, [WINDOW_SIZE[0] - 400, WINDOW_SIZE[1] - 400])
    for b in current_block:
        row = current_position[0] + b[0]
        col = current_position[1] + b[1]
        pygame.draw.rect(screen, WHITE, [(MARGIN + WIDTH) * col + MARGIN + 5, \
            (MARGIN + HEIGHT) * row + MARGIN, WIDTH, HEIGHT])

    # Draw the next block
    next_text = font.render("Next: " , True, WHITE)
    screen.blit(next_text, [WINDOW_SIZE[0] - 200, WINDOW_SIZE[1] - 400])
    for b in next_block:
	    row = next_position[0] + b[0]
	    col = next_position[1] + b[1]
	    pygame.draw.rect(screen, WHITE, [(MARGIN + WIDTH) * col + MARGIN + 5, \
            (MARGIN + HEIGHT) * row + MARGIN, WIDTH, HEIGHT])

    # Draw the score
    score_text = font.render("Score: " + str(score), True, WHITE)
    screen.blit(score_text, [WINDOW_SIZE[0] - 400, WINDOW_SIZE[1] - 120])

    # Draw the game reset button
    pygame.draw.rect(screen, button_color, reset_button_rect)
    pygame.draw.rect(screen, BLACK, reset_button_rect.inflate(10, 10))
    pygame.draw.rect(screen, WHITE, reset_button_rect, 2)
    text_surface = font.render(reset_button_text, True, text_color)
    text_rect = text_surface.get_rect(center=reset_button_rect.center)
    reset_text = font.render("RESET", True, WHITE)
    reset_button = screen.blit(reset_text, text_rect)

    # Draw undo step button
    pygame.draw.rect(screen, button_color, undo_button_rect)
    pygame.draw.rect(screen, BLACK, undo_button_rect.inflate(10, 10))
    pygame.draw.rect(screen, WHITE, undo_button_rect, 2)
    text_surface = font.render(undo_button_text, True, text_color)
    text_rect = text_surface.get_rect(center=undo_button_rect.center)
    reset_text = font.render("UNDO", True, WHITE)
    reset_button = screen.blit(reset_text, text_rect)

    # Update the screen
    pygame.display.flip()

    # Wait for the next frame
    clock.tick(60)

    # Handle events
    for event in pygame.event.get():
        if event.type == pygame.QUIT:
            done = True

        # Hanle mouse cursor movement across the game grid
        mouse_position = pygame.mouse.get_pos()
        mx, my = pygame.mouse.get_pos()
        if mx <= 410 and my <= 410:
            row = mouse_position[1] // (HEIGHT + MARGIN)
            col = mouse_position[0] // (WIDTH + MARGIN)
            draw_ghost_block(current_block, [row, col], grid, ghost_grid)

        # Handle mouse button clicks
        if event.type == pygame.MOUSEBUTTONDOWN and event.button == 1:
            mouse_position = pygame.mouse.get_pos()

            # Check if reset button was clicked
            if reset_button_rect.collidepoint(event.pos):
                reset_game()

            # Check if undo button was clicked
            if undo_button_rect.collidepoint(event.pos):
                undo_step()

            # Calculate the grid position of the mouse click
            row = mouse_position[1] // (HEIGHT + MARGIN)
            col = mouse_position[0] // (WIDTH + MARGIN)

            # Check if the current block can be placed at the clicked position
            if can_place_block(current_block, [row, col], grid):
                add_block(current_block, [row, col], grid)
                clear_lines(grid)
                prev_block = copy.deepcopy(current_block)
                current_block = next_block
                next_block = generate_block()
                score += 4

# Quit Pygame
pygame.quit()
